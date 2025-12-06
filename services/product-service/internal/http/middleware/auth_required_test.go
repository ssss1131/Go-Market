package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	mw "GoProduct/internal/http/middleware"
	jwtutil "GoProduct/pkg/jwt"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func makeToken(secret string, userID uint, email string, expiresAt time.Time) string {
	claims := jwtutil.Claims{
		UserID: userID,
		Email:  email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := tok.SignedString([]byte(secret))
	return signed
}

func TestAuthRequired_ValidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret"
	verifier := jwtutil.NewVerifier(secret)

	r := gin.New()
	r.GET("/protected", mw.AuthRequired(verifier), func(c *gin.Context) {
		raw, exists := c.Get(mw.UserIDKey)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no user id in context"})
			return
		}
		uid, ok := raw.(uint)
		if !ok || uid != 123 {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "wrong user id"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	token := makeToken(secret, 123, "user@example.com", time.Now().Add(time.Hour))
	req.Header.Set("Authorization", "Bearer "+token)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusOK, w.Body.String())
	}
}

func TestAuthRequired_MissingHeader(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret"
	verifier := jwtutil.NewVerifier(secret)

	r := gin.New()
	r.GET("/protected", mw.AuthRequired(verifier), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestAuthRequired_InvalidFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret"
	verifier := jwtutil.NewVerifier(secret)

	r := gin.New()
	r.GET("/protected", mw.AuthRequired(verifier), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Token something")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}

func TestAuthRequired_InvalidToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	secret := "test-secret"
	verifier := jwtutil.NewVerifier(secret)

	r := gin.New()
	r.GET("/protected", mw.AuthRequired(verifier), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"ok": true})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")

	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d; body = %s", w.Code, http.StatusUnauthorized, w.Body.String())
	}
}
