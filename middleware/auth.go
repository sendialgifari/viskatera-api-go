package middleware

import (
	"net/http"
	"os"
	"strings"
	"time"
	"viskatera-api-go/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"Authorization header required",
				"MISSING_AUTH_HEADER",
				"Please provide Authorization header with Bearer token",
			))
			c.Abort()
			return
		}

		// Check if token starts with "Bearer "
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"Invalid authorization header format",
				"INVALID_AUTH_FORMAT",
				"Authorization header must start with 'Bearer '",
			))
			c.Abort()
			return
		}

		// Parse and validate token
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			// Validate signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(os.Getenv("JWT_SECRET")), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"Invalid token",
				"INVALID_TOKEN",
				"Token is invalid or malformed",
			))
			c.Abort()
			return
		}

		// Extract claims
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"Invalid token claims",
				"INVALID_CLAIMS",
				"Token claims are invalid",
			))
			c.Abort()
			return
		}

		// Check token expiration
		exp, ok := claims["exp"].(float64)
		if !ok || time.Now().Unix() > int64(exp) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"Token expired",
				"TOKEN_EXPIRED",
				"Please login again to get a new token",
			))
			c.Abort()
			return
		}

		// Set user ID in context
		userID, ok := claims["user_id"].(float64)
		if !ok {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"Invalid user ID in token",
				"INVALID_USER_ID",
				"Token contains invalid user ID",
			))
			c.Abort()
			return
		}

		c.Set("user_id", uint(userID))
		c.Next()
	}
}
