package controllers

import (
	"net/http"
	"strconv"
	"viskatera-api-go/config"
	"viskatera-api-go/models"
	"viskatera-api-go/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// GetUserActivities godoc
// @Summary Get user activities
// @Description Get paginated list of all activities performed by a user or the authenticated user
// @Tags Activity
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param user_id query int false "User ID (admin only, defaults to authenticated user)"
// @Param action query string false "Filter by action (create, update, delete)"
// @Param entity_type query string false "Filter by entity type (user, visa, purchase, payment)"
// @Param page query int false "Page number (default: 1)" default(1)
// @Param per_page query int false "Items per page (default: 20, max: 100)" default(20)
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /activities [get]
func GetUserActivities(c *gin.Context) {
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to view activities",
		))
		return
	}

	// Check if user is admin and if user_id query param is provided
	var targetUserID uint = userID
	if queryUserID := c.Query("user_id"); queryUserID != "" {
		// Check if current user is admin
		var currentUser models.User
		if err := config.DB.First(&currentUser, userID).Error; err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"User not found",
				"USER_NOT_FOUND",
				"",
			))
			return
		}

		if currentUser.Role != models.RoleAdmin {
			c.JSON(http.StatusForbidden, models.ErrorResponse(
				"Access denied",
				"ACCESS_DENIED",
				"Only admins can view other users' activities",
			))
			return
		}

		// Parse target user ID
		if parsedID, err := strconv.ParseUint(queryUserID, 10, 32); err == nil {
			targetUserID = uint(parsedID)
		}
	}

	// Get query parameters
	action := c.Query("action")
	entityType := c.Query("entity_type")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	// Parse pagination
	pageInt, _ := strconv.Atoi(page)
	perPageInt, _ := strconv.Atoi(perPage)
	if pageInt < 1 {
		pageInt = 1
	}
	if perPageInt < 1 || perPageInt > 100 {
		perPageInt = 20
	}

	// Build query
	query := config.DB.Model(&models.ActivityLog{}).Where("user_id = ?", targetUserID)

	if action != "" {
		query = query.Where("action = ?", action)
	}

	if entityType != "" {
		query = query.Where("entity_type = ?", entityType)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get activities with pagination and preload user
	var activities []models.ActivityLog
	offset := (pageInt - 1) * perPageInt
	if err := query.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, email, name")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPageInt).
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch activities",
			"DATABASE_ERROR",
			"Please try again later",
		))
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(
		"Activities retrieved successfully",
		activities,
		pageInt,
		perPageInt,
		int(total),
	))
}

// GetVisaActivities godoc
// @Summary Get activities for a specific visa
// @Description Get paginated list of all activities performed on a specific visa
// @Tags Activity
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param visa_id path int true "Visa ID"
// @Param action query string false "Filter by action (create, update, delete)"
// @Param page query int false "Page number (default: 1)" default(1)
// @Param per_page query int false "Items per page (default: 20, max: 100)" default(20)
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /activities/visa/{visa_id} [get]
func GetVisaActivities(c *gin.Context) {
	visaID, err := strconv.Atoi(c.Param("visa_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid visa ID",
			"INVALID_ID",
			"Visa ID must be a valid number",
		))
		return
	}

	// Verify visa exists
	var visa models.Visa
	if err := config.DB.First(&visa, visaID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Visa not found",
			"VISA_NOT_FOUND",
			"Visa with this ID does not exist",
		))
		return
	}

	// Get query parameters
	action := c.Query("action")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	// Parse pagination
	pageInt, _ := strconv.Atoi(page)
	perPageInt, _ := strconv.Atoi(perPage)
	if pageInt < 1 {
		pageInt = 1
	}
	if perPageInt < 1 || perPageInt > 100 {
		perPageInt = 20
	}

	// Build query
	query := config.DB.Model(&models.ActivityLog{}).
		Where("entity_type = ? AND entity_id = ?", models.EntityVisa, visaID)

	if action != "" {
		query = query.Where("action = ?", action)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get activities with pagination
	var activities []models.ActivityLog
	offset := (pageInt - 1) * perPageInt
	if err := query.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, email, name")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPageInt).
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch activities",
			"DATABASE_ERROR",
			"Please try again later",
		))
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(
		"Visa activities retrieved successfully",
		activities,
		pageInt,
		perPageInt,
		int(total),
	))
}

// GetPurchaseActivities godoc
// @Summary Get activities for a specific purchase
// @Description Get paginated list of all activities performed on a specific purchase
// @Tags Activity
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param purchase_id path int true "Purchase ID"
// @Param action query string false "Filter by action (create, update, delete)"
// @Param page query int false "Page number (default: 1)" default(1)
// @Param per_page query int false "Items per page (default: 20, max: 100)" default(20)
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /activities/purchase/{purchase_id} [get]
func GetPurchaseActivities(c *gin.Context) {
	purchaseID, err := strconv.Atoi(c.Param("purchase_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid purchase ID",
			"INVALID_ID",
			"Purchase ID must be a valid number",
		))
		return
	}

	// Get authenticated user
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to view activities",
		))
		return
	}

	// Verify purchase exists and user has access
	var purchase models.VisaPurchase
	if err := config.DB.First(&purchase, purchaseID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Purchase not found",
			"PURCHASE_NOT_FOUND",
			"Purchase with this ID does not exist",
		))
		return
	}

	// Check if user is admin or owner of purchase
	var currentUser models.User
	if err := config.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not found",
			"USER_NOT_FOUND",
			"",
		))
		return
	}

	if currentUser.Role != models.RoleAdmin && purchase.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			"Access denied",
			"ACCESS_DENIED",
			"You can only view activities for your own purchases",
		))
		return
	}

	// Get query parameters
	action := c.Query("action")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	// Parse pagination
	pageInt, _ := strconv.Atoi(page)
	perPageInt, _ := strconv.Atoi(perPage)
	if pageInt < 1 {
		pageInt = 1
	}
	if perPageInt < 1 || perPageInt > 100 {
		perPageInt = 20
	}

	// Build query
	query := config.DB.Model(&models.ActivityLog{}).
		Where("entity_type = ? AND entity_id = ?", models.EntityPurchase, purchaseID)

	if action != "" {
		query = query.Where("action = ?", action)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get activities with pagination
	var activities []models.ActivityLog
	offset := (pageInt - 1) * perPageInt
	if err := query.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, email, name")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPageInt).
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch activities",
			"DATABASE_ERROR",
			"Please try again later",
		))
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(
		"Purchase activities retrieved successfully",
		activities,
		pageInt,
		perPageInt,
		int(total),
	))
}

// GetPaymentActivities godoc
// @Summary Get activities for a specific payment
// @Description Get paginated list of all activities performed on a specific payment
// @Tags Activity
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param payment_id path int true "Payment ID"
// @Param action query string false "Filter by action (create, update, delete)"
// @Param page query int false "Page number (default: 1)" default(1)
// @Param per_page query int false "Items per page (default: 20, max: 100)" default(20)
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /activities/payment/{payment_id} [get]
func GetPaymentActivities(c *gin.Context) {
	paymentID, err := strconv.Atoi(c.Param("payment_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid payment ID",
			"INVALID_ID",
			"Payment ID must be a valid number",
		))
		return
	}

	// Get authenticated user
	userID, exists := utils.GetUserIDFromContext(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to view activities",
		))
		return
	}

	// Verify payment exists and user has access
	var payment models.Payment
	if err := config.DB.First(&payment, paymentID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Payment not found",
			"PAYMENT_NOT_FOUND",
			"Payment with this ID does not exist",
		))
		return
	}

	// Check if user is admin or owner of payment
	var currentUser models.User
	if err := config.DB.First(&currentUser, userID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not found",
			"USER_NOT_FOUND",
			"",
		))
		return
	}

	if currentUser.Role != models.RoleAdmin && payment.UserID != userID {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			"Access denied",
			"ACCESS_DENIED",
			"You can only view activities for your own payments",
		))
		return
	}

	// Get query parameters
	action := c.Query("action")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "20")

	// Parse pagination
	pageInt, _ := strconv.Atoi(page)
	perPageInt, _ := strconv.Atoi(perPage)
	if pageInt < 1 {
		pageInt = 1
	}
	if perPageInt < 1 || perPageInt > 100 {
		perPageInt = 20
	}

	// Build query
	query := config.DB.Model(&models.ActivityLog{}).
		Where("entity_type = ? AND entity_id = ?", models.EntityPayment, paymentID)

	if action != "" {
		query = query.Where("action = ?", action)
	}

	// Get total count
	var total int64
	query.Count(&total)

	// Get activities with pagination
	var activities []models.ActivityLog
	offset := (pageInt - 1) * perPageInt
	if err := query.
		Preload("User", func(db *gorm.DB) *gorm.DB {
			return db.Select("id, email, name")
		}).
		Order("created_at DESC").
		Offset(offset).
		Limit(perPageInt).
		Find(&activities).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch activities",
			"DATABASE_ERROR",
			"Please try again later",
		))
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(
		"Payment activities retrieved successfully",
		activities,
		pageInt,
		perPageInt,
		int(total),
	))
}
