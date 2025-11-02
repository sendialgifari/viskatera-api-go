package utils

import (
	"encoding/json"
	"viskatera-api-go/config"
	"viskatera-api-go/models"

	"github.com/gin-gonic/gin"
)

// LogActivity creates an activity log entry
func LogActivity(c *gin.Context, userID uint, action models.ActivityAction, entityType models.ActivityEntity, entityID uint, entityName string, description string, changes interface{}) error {
	// Get IP address and user agent
	ipAddress := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Convert changes to JSON string if provided
	var changesJSON string
	if changes != nil {
		jsonBytes, err := json.Marshal(changes)
		if err == nil {
			changesJSON = string(jsonBytes)
		}
	}

	// Create activity log
	activityLog := models.ActivityLog{
		UserID:      userID,
		Action:      action,
		EntityType:  entityType,
		EntityID:    entityID,
		EntityName:  entityName,
		Description: description,
		Changes:     changesJSON,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}

	// Save in background to avoid blocking main request
	go func() {
		// Use a new database connection to avoid connection pool issues
		if err := config.DB.Create(&activityLog).Error; err != nil {
			// Log error but don't fail the main operation
			// In production, you might want to use proper logging here
			_ = err
		}
	}()

	return nil
}

// LogCreate logs a create operation
func LogCreate(c *gin.Context, userID uint, entityType models.ActivityEntity, entityID uint, entityName string, entity interface{}) error {
	description := "Created " + string(entityType) + ": " + entityName

	// Extract relevant fields for changes
	changes := map[string]interface{}{
		"action": "create",
		"entity": entity,
	}

	return LogActivity(c, userID, models.ActionCreate, entityType, entityID, entityName, description, changes)
}

// LogUpdate logs an update operation
func LogUpdate(c *gin.Context, userID uint, entityType models.ActivityEntity, entityID uint, entityName string, oldValues, newValues interface{}) error {
	description := "Updated " + string(entityType) + ": " + entityName

	changes := map[string]interface{}{
		"action":     "update",
		"old_values": oldValues,
		"new_values": newValues,
	}

	return LogActivity(c, userID, models.ActionUpdate, entityType, entityID, entityName, description, changes)
}

// LogDelete logs a delete operation
func LogDelete(c *gin.Context, userID uint, entityType models.ActivityEntity, entityID uint, entityName string, entity interface{}) error {
	description := "Deleted " + string(entityType) + ": " + entityName

	changes := map[string]interface{}{
		"action": "delete",
		"entity": entity,
	}

	return LogActivity(c, userID, models.ActionDelete, entityType, entityID, entityName, description, changes)
}

// GetUserIDFromContext extracts user ID from gin context
func GetUserIDFromContext(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}

	// Handle both uint and float64 (from JWT claims)
	switch v := userID.(type) {
	case uint:
		return v, true
	case float64:
		return uint(v), true
	default:
		return 0, false
	}
}

// GetUserIDFromContextWithDefault extracts user ID or returns system user (0)
func GetUserIDFromContextWithDefault(c *gin.Context) uint {
	userID, exists := GetUserIDFromContext(c)
	if !exists {
		return 0 // System user
	}
	return userID
}
