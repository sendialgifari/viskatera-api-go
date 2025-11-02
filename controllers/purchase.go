package controllers

import (
	"net/http"
	"strconv"
	"viskatera-api-go/config"
	"viskatera-api-go/models"
	"viskatera-api-go/utils"

	"github.com/gin-gonic/gin"
)

// UpdatePurchaseStatusRequest represents request body for updating purchase status
type UpdatePurchaseStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending completed cancelled" example:"completed"`
}

// PurchaseVisa godoc
// @Summary Purchase a visa
// @Description Create a new visa purchase. Automatically calculates total price including optional visa option.
// @Tags Purchase
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.PurchaseRequest true "Purchase request data"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /purchases [post]
func PurchaseVisa(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to purchase visa",
		))
		return
	}

	var req models.PurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request data",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	// Get visa details
	var visa models.Visa
	if err := config.DB.Where("id = ? AND is_active = ?", req.VisaID, true).First(&visa).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Visa not found or inactive",
			"VISA_NOT_FOUND",
			"Visa with this ID does not exist or is inactive",
		))
		return
	}

	// Calculate total price
	totalPrice := visa.Price

	// Check if visa option is provided and valid
	if req.VisaOptionID != nil {
		var option models.VisaOption
		if err := config.DB.Where("id = ? AND visa_id = ? AND is_active = ?", *req.VisaOptionID, req.VisaID, true).First(&option).Error; err != nil {
			c.JSON(http.StatusNotFound, models.ErrorResponse(
				"Visa option not found or inactive",
				"VISA_OPTION_NOT_FOUND",
				"Visa option with this ID does not exist or is inactive",
			))
			return
		}
		totalPrice += option.Price
	}

	// Create purchase record
	purchase := models.VisaPurchase{
		UserID:       userID.(uint),
		VisaID:       req.VisaID,
		VisaOptionID: req.VisaOptionID,
		TotalPrice:   totalPrice,
		Status:       "pending",
	}

	if err := config.DB.Create(&purchase).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create purchase",
			"PURCHASE_CREATION_ERROR",
			"Please try again later",
		))
		return
	}

	// Load related data for response
	config.DB.Preload("Visa").Preload("VisaOption").First(&purchase, purchase.ID)

	// Log activity
	logUserID := utils.GetUserIDFromContextWithDefault(c)
	entityName := "Purchase #" + strconv.Itoa(int(purchase.ID)) + " - " + purchase.Visa.Country
	utils.LogCreate(c, logUserID, models.EntityPurchase, purchase.ID, entityName, purchase)

	c.JSON(http.StatusCreated, models.SuccessResponse(
		"Visa purchase created successfully",
		purchase,
	))
}

// GetUserPurchases godoc
// @Summary Get user purchases
// @Description Get paginated list of purchases for the authenticated user
// @Tags Purchase
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number (default: 1)" default(1)
// @Param per_page query int false "Items per page (default: 10, max: 100)" default(10)
// @Success 200 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /purchases [get]
func GetUserPurchases(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to view purchases",
		))
		return
	}

	// Get pagination parameters
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "10")

	pageInt, _ := strconv.Atoi(page)
	perPageInt, _ := strconv.Atoi(perPage)
	if pageInt < 1 {
		pageInt = 1
	}
	if perPageInt < 1 || perPageInt > 100 {
		perPageInt = 10
	}

	// Get total count
	var total int64
	config.DB.Model(&models.VisaPurchase{}).Where("user_id = ?", userID).Count(&total)

	// Get purchases with pagination
	var purchases []models.VisaPurchase
	offset := (pageInt - 1) * perPageInt
	if err := config.DB.Where("user_id = ?", userID).
		Preload("Visa").Preload("VisaOption").
		Offset(offset).Limit(perPageInt).
		Find(&purchases).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch purchases",
			"DATABASE_ERROR",
			"Please try again later",
		))
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(
		"Purchases retrieved successfully",
		purchases,
		pageInt,
		perPageInt,
		int(total),
	))
}

// GetPurchaseByID godoc
// @Summary Get purchase by ID
// @Description Get detailed information about a specific purchase belonging to the authenticated user
// @Tags Purchase
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Purchase ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /purchases/{id} [get]
func GetPurchaseByID(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to view purchase",
		))
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid purchase ID",
			"INVALID_ID",
			"Purchase ID must be a valid number",
		))
		return
	}

	var purchase models.VisaPurchase
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).
		Preload("Visa").Preload("VisaOption").
		First(&purchase).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Purchase not found",
			"PURCHASE_NOT_FOUND",
			"Purchase with this ID does not exist or does not belong to you",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Purchase retrieved successfully",
		purchase,
	))
}

// UpdatePurchaseStatus godoc
// @Summary Update purchase status
// @Description Update the status of a purchase. Allowed values: pending, completed, cancelled
// @Tags Purchase
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Purchase ID"
// @Param request body UpdatePurchaseStatusRequest true "Status update data"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /purchases/{id}/status [put]
func UpdatePurchaseStatus(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to update purchase",
		))
		return
	}

	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid purchase ID",
			"INVALID_ID",
			"Purchase ID must be a valid number",
		))
		return
	}

	var req UpdatePurchaseStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request data",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	var purchase models.VisaPurchase
	if err := config.DB.Where("id = ? AND user_id = ?", id, userID).First(&purchase).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Purchase not found",
			"PURCHASE_NOT_FOUND",
			"Purchase with this ID does not exist or does not belong to you",
		))
		return
	}

	// Store old status for logging
	oldStatus := purchase.Status

	purchase.Status = req.Status
	if err := config.DB.Save(&purchase).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to update purchase",
			"PURCHASE_UPDATE_ERROR",
			"Please try again later",
		))
		return
	}

	// Load related data for response
	config.DB.Preload("Visa").Preload("VisaOption").First(&purchase, purchase.ID)

	// Log activity
	userIDVal := utils.GetUserIDFromContextWithDefault(c)
	entityName := "Purchase #" + strconv.Itoa(int(purchase.ID)) + " - " + purchase.Visa.Country
	oldValues := map[string]interface{}{"status": oldStatus}
	newValues := map[string]interface{}{"status": purchase.Status}
	utils.LogUpdate(c, userIDVal, models.EntityPurchase, purchase.ID, entityName, oldValues, newValues)

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Purchase status updated successfully",
		purchase,
	))
}
