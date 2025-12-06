package middleware

import (
	jwtutil "GoProduct/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"
const Status = "status"

func AuthRequired(verifier *jwtutil.Verifier) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization format"})
			return
		}

		token := parts[1]
		claims, err := verifier.Verify(token)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			return
		}

		c.Set(UserIDKey, claims.UserID)
		c.Set(Status, claims.Status)
		c.Next()
	}
}

func RequireActive() gin.HandlerFunc {
	return func(c *gin.Context) {
		status, _ := c.Get(Status)
		if status != "ACTIVE" {
			c.AbortWithStatusJSON(403, gin.H{"error": "account not active"})
			return
		}
		c.Next()
	}
}
