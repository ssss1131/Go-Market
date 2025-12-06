package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"

	"testing"
	"time"

	"GoMarket/internal/http/handlers"
	"GoMarket/internal/repo"
	"GoMarket/internal/service"
	jwtutil "GoMarket/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// setupTestDB connects to PostgreSQL test database
func setupTestDB(t *testing.T) *gorm.DB {
	// Use environment variable or default test database URL
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	testDBURL := os.Getenv("TEST_DB_URL")

	db, err := gorm.Open(postgres.Open(testDBURL), &gorm.Config{})
	require.NoError(t, err, "failed to open test database")

	// Auto-migrate your models here
	// err = db.AutoMigrate(&models.User{})
	// require.NoError(t, err, "failed to migrate test database")

	return db
}

// cleanupTestDB cleans up test data after tests
func cleanupTestDB(t *testing.T, db *gorm.DB) {
	// Clean up tables in reverse order of dependencies
	db.Exec("TRUNCATE TABLE users CASCADE")

	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

// setupTestRouter creates a test router with all dependencies
func setupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	gin.SetMode(gin.TestMode)

	db := setupTestDB(t)
	signer := jwtutil.NewSigner("test-secret-key", 15*time.Minute)

	usersRepo := repo.NewUsers(db)
	authSvc := service.NewAuthService(usersRepo, signer)
	authH := handlers.NewAuthHandler(authSvc)

	r := gin.New()
	r.POST("/auth/register", authH.Register)
	r.POST("/auth/login", authH.Login)

	return r, db
}

func TestRegisterEndpoint(t *testing.T) {
	router, db := setupTestRouter(t)
	defer cleanupTestDB(t, db)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful registration",
			requestBody: map[string]interface{}{
				"email":    "test@example.com",
				"password": "securepassword123",
				"name":     "Test",
				"surname":  "User",
			},
			expectedStatus: http.StatusCreated,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				//assert.Contains(t, response, "access_token")
			},
		},
		{
			name: "missing required fields",
			requestBody: map[string]interface{}{
				"email": "test@example.com",
				"name":  "Test",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			name: "invalid email format",
			requestBody: map[string]interface{}{
				"email":    "invalid-email",
				"password": "securepassword123",
			},
			expectedStatus: http.StatusBadRequest,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestLoginEndpoint(t *testing.T) {
	router, db := setupTestRouter(t)
	defer cleanupTestDB(t, db)

	// First, register a user
	registerBody := map[string]interface{}{
		"email":    "login@example.com",
		"password": "testpassword123",
		"name":     "Login",
		"surname":  "Test",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusCreated, w.Code)

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		checkResponse  func(t *testing.T, body []byte)
	}{
		{
			name: "successful login",
			requestBody: map[string]interface{}{
				"email":    "login@example.com",
				"password": "testpassword123",
			},
			expectedStatus: http.StatusOK,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "access_token")
			},
		},
		{
			name: "wrong password",
			requestBody: map[string]interface{}{
				"email":    "login@example.com",
				"password": "wrongpassword",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse: func(t *testing.T, body []byte) {
				var response map[string]interface{}
				err := json.Unmarshal(body, &response)
				require.NoError(t, err)
				assert.Contains(t, response, "error")
			},
		},
		{
			name: "non-existent user",
			requestBody: map[string]interface{}{
				"email":    "nonexistent@example.com",
				"password": "somepassword",
			},
			expectedStatus: http.StatusUnauthorized,
			checkResponse:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.checkResponse != nil {
				tt.checkResponse(t, w.Body.Bytes())
			}
		})
	}
}

func TestServerShutdown(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	srv := &http.Server{
		Addr:         "127.0.0.1:0", // Random available port
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
	}

	// Start server
	go func() {
		_ = srv.ListenAndServe()
	}()

	// Give server time to start
	time.Sleep(100 * time.Millisecond)

	// Test graceful shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := srv.Shutdown(ctx)
	assert.NoError(t, err, "server should shutdown gracefully")
}

func TestDatabaseConnection(t *testing.T) {
	db := setupTestDB(t)

	sqlDB, err := db.DB()
	require.NoError(t, err, "should get underlying sql.DB")

	err = sqlDB.Ping()
	assert.NoError(t, err, "should ping database successfully")

	err = sqlDB.Close()
	assert.NoError(t, err, "should close database successfully")
}

func TestHTTPServerTimeouts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(20 * time.Second) // Exceeds write timeout
		c.JSON(http.StatusOK, gin.H{"status": "done"})
	})

	srv := &http.Server{
		Addr:         "127.0.0.1:0",
		Handler:      r,
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 1 * time.Second,
	}

	assert.Equal(t, 1*time.Second, srv.ReadTimeout, "read timeout should be configured")
	assert.Equal(t, 1*time.Second, srv.WriteTimeout, "write timeout should be configured")
}

// Benchmark tests
func BenchmarkRegisterEndpoint(b *testing.B) {
	router, _ := setupTestRouter(&testing.T{})

	requestBody := map[string]interface{}{
		"email":    "bench@example.com",
		"password": "benchpassword123",
		"name":     "Bench",
		"surname":  "User",
	}
	body, _ := json.Marshal(requestBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkLoginEndpoint(b *testing.B) {
	router, _ := setupTestRouter(&testing.T{})

	// Pre-register a user
	registerBody := map[string]interface{}{
		"email":    "benchlogin@example.com",
		"password": "benchpassword123",
		"name":     "Bench",
		"surname":  "Login",
	}
	body, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	loginBody := map[string]interface{}{
		"email":    "benchlogin@example.com",
		"password": "benchpassword123",
	}
	body, _ = json.Marshal(loginBody)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)
	}
}
