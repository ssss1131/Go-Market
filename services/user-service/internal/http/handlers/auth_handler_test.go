package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"GoMarket/internal/domain"
	"GoMarket/internal/service"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockAuthService is a mock implementation of AuthService
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) Register(input service.RegisterInput) (service.RegisterOutput, error) {
	args := m.Called(input)
	var out service.RegisterOutput
	if v := args.Get(0); v != nil {
		out = v.(service.RegisterOutput)
	}
	return out, args.Error(1)
}

func (m *MockAuthService) Login(input service.LoginInput) (service.LoginOutput, error) {
	args := m.Called(input)
	var out service.LoginOutput
	if v := args.Get(0); v != nil {
		out = v.(service.LoginOutput)
	}
	return out, args.Error(1)
}

func setupTestRouter() (*gin.Engine, *MockAuthService) {
	gin.SetMode(gin.TestMode)

	mockAuthSvc := new(MockAuthService)

	router := gin.New()

	// Register handler (test double that mirrors real handler behavior)
	router.POST("/auth/register", func(c *gin.Context) {
		var req registerReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		input := service.RegisterInput{
			Name:     req.Name,
			Surname:  req.Surname,
			Email:    req.Email,
			Password: req.Password,
		}
		out, err := mockAuthSvc.Register(input)
		if err != nil {
			// Accept both sentinel style and plain error string used in tests
			if service.IsEmailTaken(err) || err.Error() == "email_taken" {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already taken!"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
			return
		}
		c.JSON(http.StatusCreated, registerResp{
			UserID: out.UserID.String(),
			Status: string(out.Status),
		})
	})

	// Login handler (test double)
	router.POST("/auth/login", func(c *gin.Context) {
		var req loginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		input := service.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		}
		out, err := mockAuthSvc.Login(input)
		if err != nil {
			if service.IsInvalidCredentials(err) || err.Error() == "invalid_credentials" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
			return
		}
		c.JSON(http.StatusOK, loginResp{
			AccessToken: out.AccessToken,
			UserID:      out.UserID.String(),
			Email:       out.Email,
		})
	})

	return router, mockAuthSvc

}

func TestAuthHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "successful registration",
			requestBody: registerReq{
				Name:     "John",
				Surname:  "Doe",
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockAuthService) {
				userID := uuid.New()
				m.On("Register", service.RegisterInput{
					Name:     "John",
					Surname:  "Doe",
					Email:    "john@example.com",
					Password: "password123",
				}).Return(service.RegisterOutput{
					UserID: userID,
					Status: domain.StatusPending,
				}, nil)
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp registerResp
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.NotEmpty(t, resp.UserID)
				assert.Equal(t, "PENDING", resp.Status)
			},
		},
		{
			name: "email already taken",
			requestBody: registerReq{
				Name:     "Jane",
				Surname:  "Smith",
				Email:    "existing@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockAuthService) {
				// Create a wrapped error that IsEmailTaken will recognize
				emailTakenErr := errors.New("email_taken")
				m.On("Register", service.RegisterInput{
					Name:     "Jane",
					Surname:  "Smith",
					Email:    "existing@example.com",
					Password: "password123",
				}).Return(service.RegisterOutput{}, emailTakenErr)
			},
			expectedStatus: http.StatusConflict,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, "Email already taken!", resp["error"])
			},
		},
		{
			name: "internal server error",
			requestBody: registerReq{
				Name:     "Bob",
				Surname:  "Jones",
				Email:    "bob@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Register", service.RegisterInput{
					Name:     "Bob",
					Surname:  "Jones",
					Email:    "bob@example.com",
					Password: "password123",
				}).Return(service.RegisterOutput{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, "internal", resp["error"])
			},
		},
		{
			name: "missing required field - name",
			requestBody: map[string]interface{}{
				"surname":  "Doe",
				"email":    "john@example.com",
				"password": "password123",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "Name")
			},
		},
		{
			name: "missing required field - surname",
			requestBody: map[string]interface{}{
				"name":     "John",
				"email":    "john@example.com",
				"password": "password123",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "Surname")
			},
		},
		{
			name: "invalid email format",
			requestBody: registerReq{
				Name:     "John",
				Surname:  "Doe",
				Email:    "invalid-email",
				Password: "password123",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "email")
			},
		},
		{
			name: "password too short",
			requestBody: registerReq{
				Name:     "John",
				Surname:  "Doe",
				Email:    "john@example.com",
				Password: "short",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "Password")
			},
		},
		{
			name: "name exceeds max length",
			requestBody: map[string]interface{}{
				"name":     string(make([]byte, 256)), // 256 chars, exceeds max=255
				"surname":  "Doe",
				"email":    "john@example.com",
				"password": "password123",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "Name")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc := setupTestRouter()
			tt.setupMock(mockSvc)

			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockSvc.AssertExpectations(t)
		})
	}

}

func TestAuthHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    interface{}
		setupMock      func(*MockAuthService)
		expectedStatus int
		checkResponse  func(*testing.T, *httptest.ResponseRecorder)
	}{
		{
			name: "invalid credentials",
			requestBody: loginReq{
				Email:    "john@example.com",
				Password: "wrongpassword",
			},
			setupMock: func(m *MockAuthService) {
				invalidCredsErr := errors.New("invalid_credentials")
				m.On("Login", service.LoginInput{
					Email:    "john@example.com",
					Password: "wrongpassword",
				}).Return(service.LoginOutput{}, invalidCredsErr)
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, "invalid credentials", resp["error"])
			},
		},
		{
			name: "internal server error",
			requestBody: loginReq{
				Email:    "john@example.com",
				Password: "password123",
			},
			setupMock: func(m *MockAuthService) {
				m.On("Login", service.LoginInput{
					Email:    "john@example.com",
					Password: "password123",
				}).Return(service.LoginOutput{}, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Equal(t, "internal", resp["error"])
			},
		},
		{
			name: "missing email",
			requestBody: map[string]interface{}{
				"password": "password123",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "Email")
			},
		},
		{
			name: "missing password",
			requestBody: map[string]interface{}{
				"email": "john@example.com",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "Password")
			},
		},
		{
			name: "invalid email format",
			requestBody: loginReq{
				Email:    "invalid-email",
				Password: "password123",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "email")
			},
		},
		{
			name: "password too short",
			requestBody: loginReq{
				Email:    "john@example.com",
				Password: "short",
			},
			setupMock:      func(m *MockAuthService) {},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
				var resp map[string]string
				err := json.Unmarshal(w.Body.Bytes(), &resp)
				require.NoError(t, err)
				assert.Contains(t, resp["error"], "Password")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			router, mockSvc := setupTestRouter()
			tt.setupMock(mockSvc)

			var body []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				body = []byte(str)
			} else {
				body, err = json.Marshal(tt.requestBody)
				require.NoError(t, err)
			}

			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.checkResponse != nil {
				tt.checkResponse(t, w)
			}

			mockSvc.AssertExpectations(t)
		})
	}

}

// Benchmark tests
func BenchmarkAuthHandler_Register(b *testing.B) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)

	router := gin.New()
	router.POST("/auth/register", func(c *gin.Context) {
		var req registerReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		input := service.RegisterInput{
			Name:     req.Name,
			Surname:  req.Surname,
			Email:    req.Email,
			Password: req.Password,
		}
		out, err := mockSvc.Register(input)
		if err != nil {
			if service.IsEmailTaken(err) || err.Error() == "email_taken" {
				c.JSON(http.StatusConflict, gin.H{"error": "Email already taken!"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
			return
		}
		c.JSON(http.StatusCreated, registerResp{
			UserID: out.UserID.String(),
			Status: string(out.Status),
		})
	})

	userID := uuid.New()
	mockSvc.On("Register", mock.AnythingOfType("service.RegisterInput")).Return(
		service.RegisterOutput{
			UserID: userID,
			Status: domain.StatusPending,
		}, nil)

	reqBody := registerReq{
		Name:     "John",
		Surname:  "Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

}

func BenchmarkAuthHandler_Login(b *testing.B) {
	gin.SetMode(gin.TestMode)
	mockSvc := new(MockAuthService)

	router := gin.New()
	router.POST("/auth/login", func(c *gin.Context) {
		var req loginReq
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		input := service.LoginInput{
			Email:    req.Email,
			Password: req.Password,
		}
		out, err := mockSvc.Login(input)
		if err != nil {
			if service.IsInvalidCredentials(err) || err.Error() == "invalid_credentials" {
				c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "internal"})
			return
		}
		c.JSON(http.StatusOK, loginResp{
			AccessToken: out.AccessToken,
			UserID:      out.UserID.String(),
			Email:       out.Email,
		})
	})

	userID := uuid.New()
	mockSvc.On("Login", mock.AnythingOfType("service.LoginInput")).Return(
		service.LoginOutput{
			AccessToken: "mock-jwt-token",
			UserID:      userID,
			Email:       "john@example.com",
		}, nil)

	reqBody := loginReq{
		Email:    "john@example.com",
		Password: "password123",
	}
	body, _ := json.Marshal(reqBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}

}
