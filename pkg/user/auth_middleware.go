package user

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/huinong/golang-claude/pkg/common/errors"
	"github.com/huinong/golang-claude/pkg/common/response"
)

// AuthMiddleware creates a Gin authentication middleware using API Key.
func AuthMiddleware(service Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		apiKey := c.GetHeader("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}
		if apiKey == "" || strings.TrimSpace(apiKey) == "" {
			response.Error(c, errors.New(errors.CodeUnauthorized, "missing X-API-Key header"))
			c.Abort()
			return
		}

		user, err := service.GetByAPIKey(apiKey)
		if err != nil {
			response.Error(c, err)
			c.Abort()
			return
		}

		c.Set("user_id", user.ID)
		c.Next()
	}
}

// GetUserID extracts the authenticated user ID from Gin context.
func GetUserID(c *gin.Context) uint {
	id, exists := c.Get("user_id")
	if !exists {
		return 0
	}
	return id.(uint)
}
