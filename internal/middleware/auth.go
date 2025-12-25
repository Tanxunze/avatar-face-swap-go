package middleware

import (
	"strings"

	"avatar-face-swap-go/internal/service"
	"avatar-face-swap-go/pkg/response"

	"github.com/gin-gonic/gin"
)

func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, 401, "Missing authorization header")
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, 401, "Invalid authorization format")
			c.Abort()
			return
		}

		claims, err := service.ValidateJWT(parts[1])
		if err != nil {
			response.Error(c, 401, "Invalid token: "+err.Error())
			c.Abort()
			return
		}

		// Store claims in context for handlers to use
		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)
		c.Set("user_email", claims.UserEmail)

		c.Next()
	}
}

func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists || role != "admin" {
			response.Error(c, 403, "Admin permission required")
			c.Abort()
			return
		}
		c.Next()
	}
}

func EventPermission() gin.HandlerFunc {
	return func(c *gin.Context) {
		role, _ := c.Get("role")
		eventID := c.Param("id")

		// Admin can access all events
		if role == "admin" {
			c.Next()
			return
		}

		// User can only access their authorized event
		if role != eventID {
			response.Error(c, 403, "No permission to access this event")
			c.Abort()
			return
		}

		c.Next()
	}
}
