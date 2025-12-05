package service

import (
	"GoUser/internal/domain"
	"GoUser/internal/repo"
	jwtutil "GoUser/pkg/jwt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	dsn := "host=localhost user=test password=test dbname=test_db port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}

func setupTestService(t *testing.T) (*AuthService, *gorm.DB, func()) {
	db := setupTestDB(t)
	err := db.AutoMigrate(&domain.User{})
	require.NoError(t, err)

	users := repo.NewUsers(db)
	signer := jwtutil.NewSigner("test-secret-key", time.Hour)
	service := NewAuthService(users, signer)

	cleanup := func() {
		db.Exec("DELETE FROM users")
	}

	return service, db, cleanup
}

func TestAuthService_Register_Success(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	input := RegisterInput{
		Name:     "John",
		Surname:  "Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	output, err := service.Register(input)

	assert.NoError(t, err)
	assert.NotZero(t, output.UserID)
	assert.Equal(t, domain.StatusPending, output.Status)
}

func TestAuthService_Register_EmailTaken(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	input := RegisterInput{
		Name:     "John",
		Surname:  "Doe",
		Email:    "john@example.com",
		Password: "password123",
	}

	_, err := service.Register(input)
	require.NoError(t, err)

	output, err := service.Register(input)

	assert.Error(t, err)
	assert.True(t, IsEmailTaken(err))
	assert.Equal(t, uint(0), output.UserID)
}

func TestAuthService_Register_EmptyFields(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	testCases := []struct {
		name  string
		input RegisterInput
	}{
		{
			name: "empty email",
			input: RegisterInput{
				Name:     "John",
				Surname:  "Doe",
				Email:    "",
				Password: "password123",
			},
		},
		{
			name: "empty password",
			input: RegisterInput{
				Name:     "John",
				Surname:  "Doe",
				Email:    "john@example.com",
				Password: "",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			output, err := service.Register(tc.input)

			if err == nil {
				assert.NotZero(t, output.UserID)
			}
		})
	}
}

func TestAuthService_Register_MultipleUsers(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	users := []RegisterInput{
		{Name: "John", Surname: "Doe", Email: "john@example.com", Password: "pass1"},
		{Name: "Jane", Surname: "Smith", Email: "jane@example.com", Password: "pass2"},
		{Name: "Bob", Surname: "Johnson", Email: "bob@example.com", Password: "pass3"},
	}

	for _, input := range users {
		output, err := service.Register(input)
		assert.NoError(t, err)
		assert.NotZero(t, output.UserID)
		assert.Equal(t, domain.StatusPending, output.Status)
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	registerInput := RegisterInput{
		Name:     "John",
		Surname:  "Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	regOutput, err := service.Register(registerInput)
	require.NoError(t, err)

	loginInput := LoginInput{
		Email:    "john@example.com",
		Password: "password123",
	}

	output, err := service.Login(loginInput)

	assert.NoError(t, err)
	assert.NotEmpty(t, output.AccessToken)
	assert.Equal(t, regOutput.UserID, output.UserID)
	assert.Equal(t, "john@example.com", output.Email)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	input := LoginInput{
		Email:    "nonexistent@example.com",
		Password: "password123",
	}

	output, err := service.Login(input)

	assert.Error(t, err)
	assert.True(t, IsInvalidCredentials(err))
	assert.Empty(t, output.AccessToken)
	assert.Equal(t, uint(0), output.UserID)
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	registerInput := RegisterInput{
		Name:     "John",
		Surname:  "Doe",
		Email:    "john@example.com",
		Password: "correctpassword",
	}
	_, err := service.Register(registerInput)
	require.NoError(t, err)

	loginInput := LoginInput{
		Email:    "john@example.com",
		Password: "wrongpassword",
	}

	output, err := service.Login(loginInput)

	assert.Error(t, err)
	assert.True(t, IsInvalidCredentials(err))
	assert.Empty(t, output.AccessToken)
}

func TestAuthService_Login_EmptyPassword(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	registerInput := RegisterInput{
		Name:     "John",
		Surname:  "Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	_, err := service.Register(registerInput)
	require.NoError(t, err)

	loginInput := LoginInput{
		Email:    "john@example.com",
		Password: "",
	}

	output, err := service.Login(loginInput)

	assert.Error(t, err)
	assert.True(t, IsInvalidCredentials(err))
	assert.Empty(t, output.AccessToken)
}

func TestAuthService_Login_CaseSensitiveEmail(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	registerInput := RegisterInput{
		Name:     "John",
		Surname:  "Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	_, err := service.Register(registerInput)
	require.NoError(t, err)

	loginInput := LoginInput{
		Email:    "John@Example.COM",
		Password: "password123",
	}

	output, err := service.Login(loginInput)

	if err != nil {
		assert.True(t, IsInvalidCredentials(err))
	} else {
		assert.NotEmpty(t, output.AccessToken)
	}
}

func TestAuthService_Login_TokenIsValid(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	registerInput := RegisterInput{
		Name:     "John",
		Surname:  "Doe",
		Email:    "john@example.com",
		Password: "password123",
	}
	regOutput, err := service.Register(registerInput)
	require.NoError(t, err)

	loginInput := LoginInput{
		Email:    "john@example.com",
		Password: "password123",
	}
	loginOutput, err := service.Login(loginInput)
	require.NoError(t, err)

	claims, err := service.signer.Verify(loginOutput.AccessToken)

	assert.NoError(t, err)
	assert.Equal(t, regOutput.UserID, claims.UserID)
	assert.Equal(t, "john@example.com", claims.Email)
}

func TestAuthService_Register_Login_Workflow(t *testing.T) {
	service, _, cleanup := setupTestService(t)
	defer cleanup()

	email := "workflow@example.com"
	password := "securepass123"

	regOutput, err := service.Register(RegisterInput{
		Name:     "Workflow",
		Surname:  "Test",
		Email:    email,
		Password: password,
	})
	assert.NoError(t, err)
	assert.NotZero(t, regOutput.UserID)
	assert.Equal(t, domain.StatusPending, regOutput.Status)

	loginOutput, err := service.Login(LoginInput{
		Email:    email,
		Password: password,
	})
	assert.NoError(t, err)
	assert.NotEmpty(t, loginOutput.AccessToken)
	assert.Equal(t, regOutput.UserID, loginOutput.UserID)
	assert.Equal(t, email, loginOutput.Email)

	claims, err := service.signer.Verify(loginOutput.AccessToken)
	assert.NoError(t, err)
	assert.Equal(t, regOutput.UserID, claims.UserID)
	assert.Equal(t, email, claims.Email)
}

func TestIsEmailTaken(t *testing.T) {
	assert.True(t, IsEmailTaken(errEmailTaken))
	assert.False(t, IsEmailTaken(errInvalidCredentials))
	assert.False(t, IsEmailTaken(nil))
}

func TestIsInvalidCredentials(t *testing.T) {
	assert.True(t, IsInvalidCredentials(errInvalidCredentials))
	assert.False(t, IsInvalidCredentials(errEmailTaken))
	assert.False(t, IsInvalidCredentials(nil))
}
