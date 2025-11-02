package controllers

import (
	"net/http"
	"strconv"
	"viskatera-api-go/config"
	"viskatera-api-go/models"
	"viskatera-api-go/utils"

	"github.com/gin-gonic/gin"
)

// CreateVisaRequest represents request body for creating a visa
type CreateVisaRequest struct {
	Country     string  `json:"country" binding:"required" example:"Japan"`
	Type        string  `json:"type" binding:"required" example:"Tourist"`
	Description string  `json:"description" example:"Tourist visa for Japan with 30 days validity"`
	Price       float64 `json:"price" binding:"required,min=0" example:"500000"`
	Duration    int     `json:"duration" binding:"required,min=1" example:"30"`
	IsActive    bool    `json:"is_active" example:"true"`
}

// UpdateVisaRequest represents request body for updating a visa
type UpdateVisaRequest struct {
	Country     string  `json:"country" example:"Japan"`
	Type        string  `json:"type" example:"Business"`
	Description string  `json:"description" example:"Updated description"`
	Price       float64 `json:"price" binding:"min=0" example:"600000"`
	Duration    int     `json:"duration" binding:"min=1" example:"60"`
	IsActive    *bool   `json:"is_active" example:"true"`
}

// GetVisas godoc
// @Summary Get all visas
// @Description Get list of all active visas with optional filtering and pagination
// @Tags Visa
// @Accept json
// @Produce json
// @Param country query string false "Filter by country (case-insensitive partial match)"
// @Param type query string false "Filter by visa type (case-insensitive partial match)"
// @Param page query int false "Page number (default: 1)" default(1)
// @Param per_page query int false "Items per page (default: 10, max: 100)" default(10)
// @Success 200 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /visas [get]
func GetVisas(c *gin.Context) {
	var visas []models.Visa

	// Get query parameters
	country := c.Query("country")
	visaType := c.Query("type")
	page := c.DefaultQuery("page", "1")
	perPage := c.DefaultQuery("per_page", "10")

	// Parse pagination
	pageInt, _ := strconv.Atoi(page)
	perPageInt, _ := strconv.Atoi(perPage)
	if pageInt < 1 {
		pageInt = 1
	}
	if perPageInt < 1 || perPageInt > 100 {
		perPageInt = 10
	}

	query := config.DB.Where("is_active = ?", true)

	if country != "" {
		query = query.Where("country ILIKE ?", "%"+country+"%")
	}

	if visaType != "" {
		query = query.Where("type ILIKE ?", "%"+visaType+"%")
	}

	// Get total count
	var total int64
	query.Model(&models.Visa{}).Count(&total)

	// Apply pagination
	offset := (pageInt - 1) * perPageInt
	if err := query.Offset(offset).Limit(perPageInt).Find(&visas).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to fetch visas",
			"DATABASE_ERROR",
			"Please try again later",
		))
		return
	}

	c.JSON(http.StatusOK, models.PaginatedResponse(
		"Visas retrieved successfully",
		visas,
		pageInt,
		perPageInt,
		int(total),
	))
}

// GetVisaByID godoc
// @Summary Get visa by ID
// @Description Get detailed information about a specific visa including its options
// @Tags Visa
// @Accept json
// @Produce json
// @Param id path int true "Visa ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Router /visas/{id} [get]
func GetVisaByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid visa ID",
			"INVALID_ID",
			"Visa ID must be a valid number",
		))
		return
	}

	var visa models.Visa
	if err := config.DB.Where("id = ? AND is_active = ?", id, true).First(&visa).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Visa not found",
			"VISA_NOT_FOUND",
			"Visa with this ID does not exist or is inactive",
		))
		return
	}

	// Get visa options
	var options []models.VisaOption
	config.DB.Where("visa_id = ? AND is_active = ?", visa.ID, true).Find(&options)

	responseData := gin.H{
		"visa":    visa,
		"options": options,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Visa retrieved successfully",
		responseData,
	))
}

// CreateVisa godoc
// @Summary Create new visa
// @Description Create a new visa (admin only)
// @Tags Visa
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateVisaRequest true "Visa creation data"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /admin/visas [post]
func CreateVisa(c *gin.Context) {
	var req CreateVisaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request data",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	visa := models.Visa{
		Country:     req.Country,
		Type:        req.Type,
		Description: req.Description,
		Price:       req.Price,
		Duration:    req.Duration,
		IsActive:    req.IsActive,
	}

	if err := config.DB.Create(&visa).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create visa",
			"VISA_CREATION_ERROR",
			"Please try again later",
		))
		return
	}

	// Log activity
	userID := utils.GetUserIDFromContextWithDefault(c)
	entityName := visa.Country + " - " + visa.Type
	utils.LogCreate(c, userID, models.EntityVisa, visa.ID, entityName, visa)

	c.JSON(http.StatusCreated, models.SuccessResponse(
		"Visa created successfully",
		visa,
	))
}

// UpdateVisa godoc
// @Summary Update visa
// @Description Update an existing visa (admin only). All fields are optional.
// @Tags Visa
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Visa ID"
// @Param request body UpdateVisaRequest true "Visa update data"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /admin/visas/{id} [put]
func UpdateVisa(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid visa ID",
			"INVALID_ID",
			"Visa ID must be a valid number",
		))
		return
	}

	var visa models.Visa
	if err := config.DB.First(&visa, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Visa not found",
			"VISA_NOT_FOUND",
			"Visa with this ID does not exist",
		))
		return
	}

	// Store old values for logging
	oldValues := map[string]interface{}{
		"country":     visa.Country,
		"type":        visa.Type,
		"description": visa.Description,
		"price":       visa.Price,
		"duration":    visa.Duration,
		"is_active":   visa.IsActive,
	}

	var req UpdateVisaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request data",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	// Update fields if provided
	if req.Country != "" {
		visa.Country = req.Country
	}
	if req.Type != "" {
		visa.Type = req.Type
	}
	if req.Description != "" {
		visa.Description = req.Description
	}
	if req.Price > 0 {
		visa.Price = req.Price
	}
	if req.Duration > 0 {
		visa.Duration = req.Duration
	}
	if req.IsActive != nil {
		visa.IsActive = *req.IsActive
	}

	if err := config.DB.Save(&visa).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to update visa",
			"VISA_UPDATE_ERROR",
			"Please try again later",
		))
		return
	}

	// Store new values for logging
	newValues := map[string]interface{}{
		"country":     visa.Country,
		"type":        visa.Type,
		"description": visa.Description,
		"price":       visa.Price,
		"duration":    visa.Duration,
		"is_active":   visa.IsActive,
	}

	// Log activity
	userID := utils.GetUserIDFromContextWithDefault(c)
	entityName := visa.Country + " - " + visa.Type
	utils.LogUpdate(c, userID, models.EntityVisa, visa.ID, entityName, oldValues, newValues)

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Visa updated successfully",
		visa,
	))
}

// DeleteVisa godoc
// @Summary Delete visa
// @Description Delete a visa (admin only). Uses soft delete.
// @Tags Visa
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Visa ID"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 403 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /admin/visas/{id} [delete]
func DeleteVisa(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid visa ID",
			"INVALID_ID",
			"Visa ID must be a valid number",
		))
		return
	}

	var visa models.Visa
	if err := config.DB.First(&visa, id).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Visa not found",
			"VISA_NOT_FOUND",
			"Visa with this ID does not exist",
		))
		return
	}

	// Log activity before delete
	userID := utils.GetUserIDFromContextWithDefault(c)
	entityName := visa.Country + " - " + visa.Type
	utils.LogDelete(c, userID, models.EntityVisa, visa.ID, entityName, visa)

	if err := config.DB.Delete(&visa).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to delete visa",
			"VISA_DELETE_ERROR",
			"Please try again later",
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Visa deleted successfully",
		nil,
	))
}
