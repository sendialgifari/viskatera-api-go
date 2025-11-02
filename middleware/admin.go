package middleware

import (
	"net/http"
	"viskatera-api-go/config"
	"viskatera-api-go/models"

	"github.com/gin-gonic/gin"
)

func AdminMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user ID from context (set by AuthMiddleware)
		userID, exists := c.Get("user_id")
		if !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"User not authenticated",
				"UNAUTHORIZED",
				"Authentication required",
			))
			c.Abort()
			return
		}

		// Get user from database
		var user models.User
		if err := config.DB.Where("id = ? AND is_active = ?", userID, true).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"User not found or inactive",
				"USER_NOT_FOUND",
				"User account not found or deactivated",
			))
			c.Abort()
			return
		}

		// Check if user is admin
		if user.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				"Access denied",
				"ACCESS_DENIED",
				"Admin privileges required",
			))
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user", user)
		c.Next()
	}
}
