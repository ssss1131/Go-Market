package middleware

import (
	jwtutil "GoProduct/pkg/jwt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const UserIDKey = "user_id"

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
		c.Next()
	}
}
