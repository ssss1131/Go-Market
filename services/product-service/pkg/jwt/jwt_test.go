package jwt

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func generateToken(secret string, claims jwt.Claims) string {
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	token, _ := t.SignedString([]byte(secret))
	return token
}

func TestVerifier_Verify_Success(t *testing.T) {
	secret := "mysecret"
	v := NewVerifier(secret)

	claims := &Claims{
		UserID: 42,
		Email:  "john@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}

	token := generateToken(secret, claims)

	out, err := v.Verify(token)

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, uint(42), out.UserID)
	assert.Equal(t, "john@example.com", out.Email)
}

func TestVerifier_Verify_InvalidToken(t *testing.T) {
	v := NewVerifier("secret")

	out, err := v.Verify("not-a-token")

	assert.Error(t, err)
	assert.Nil(t, out)
}

func TestVerifier_Verify_WrongSecret(t *testing.T) {
	claims := &Claims{
		UserID: 1,
		Email:  "a@b.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute)),
		},
	}

	token := generateToken("secret1", claims)

	v := NewVerifier("secret2")

	out, err := v.Verify(token)

	assert.Error(t, err)
	assert.Nil(t, out)
}

func TestVerifier_Verify_Expired(t *testing.T) {
	secret := "abc"
	v := NewVerifier(secret)

	claims := &Claims{
		UserID: 99,
		Email:  "expired@mail.com",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Minute)),
		},
	}

	token := generateToken(secret, claims)

	out, err := v.Verify(token)

	assert.Error(t, err)
	assert.Nil(t, out)
	assert.Contains(t, err.Error(), "expired")
}

func TestVerifier_Verify_InvalidClaims(t *testing.T) {
	v := NewVerifier("secret")

	token := generateToken("secret", jwt.MapClaims{
		"foo": "bar",
	})

	out, err := v.Verify(token)

	assert.NoError(t, err)
	assert.NotNil(t, out)
	assert.Equal(t, uint(0), out.UserID)
	assert.Equal(t, "", out.Email)
}
