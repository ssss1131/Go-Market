package jwt_test

import (
	"testing"
	"time"

	jwtpkg "GoProduct/pkg/jwt"

	"github.com/golang-jwt/jwt/v5"
)

func makeToken(t *testing.T, secret string, userID uint, email string, expiresAt time.Time) string {
	t.Helper()

	claims := jwtpkg.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("SignedString() error = %v", err)
	}
	return signed
}

func TestVerifier_Verify_ValidToken(t *testing.T) {
	secret := "test-secret"
	ver := jwtpkg.NewVerifier(secret)

	tokenStr := makeToken(t, secret, 7, "user@example.com", time.Now().Add(1*time.Hour))

	claims, err := ver.Verify(tokenStr)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}

	if claims.UserID != 7 {
		t.Errorf("UserID = %d, want %d", claims.UserID, 7)
	}
	if claims.Email != "user@example.com" {
		t.Errorf("Email = %q, want %q", claims.Email, "user@example.com")
	}
}

func TestVerifier_Verify_ExpiredToken(t *testing.T) {
	secret := "test-secret"
	ver := jwtpkg.NewVerifier(secret)

	tokenStr := makeToken(t, secret, 1, "expired@example.com", time.Now().Add(-1*time.Hour))

	_, err := ver.Verify(tokenStr)
	if err == nil {
		t.Fatalf("expected error for expired token, got nil")
	}
}

func TestVerifier_Verify_InvalidSignature(t *testing.T) {
	ver := jwtpkg.NewVerifier("right-secret")

	tokenStr := makeToken(t, "wrong-secret", 1, "user@example.com", time.Now().Add(1*time.Hour))

	_, err := ver.Verify(tokenStr)
	if err == nil {
		t.Fatalf("expected error for invalid signature, got nil")
	}
}
