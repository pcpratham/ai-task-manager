package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/task-manager/backend/internal/service"
)

const (
	UserIDKey = "user_id"
	EmailKey  = "email"
)

func AuthMiddleware(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Authorization header is required",
			})
			return
		}

		// Check for Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid authorization header format. Use: Bearer <token>",
			})
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := authService.ValidateJWT(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid or expired token",
			})
			return
		}

		// Set user info in context
		c.Set(UserIDKey, claims.UserID)
		c.Set(EmailKey, claims.Email)

		c.Next()
	}
}

// GetUserID extracts user ID from context
func GetUserID(c *gin.Context) string {
	userID, exists := c.Get(UserIDKey)
	if !exists {
		return ""
	}
	return userID.(string)
}

// GetEmail extracts email from context
func GetEmail(c *gin.Context) string {
	email, exists := c.Get(EmailKey)
	if !exists {
		return ""
	}
	return email.(string)
}
