package controllers

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"viskatera-api-go/config"
	"viskatera-api-go/models"
	"viskatera-api-go/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// RequestOTP godoc
// @Summary Request OTP for login
// @Description Generates and sends OTP code to user's email for login
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RequestOTPRequest true "Email address"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/request-otp [post]
func RequestOTP(c *gin.Context) {
	var req models.RequestOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request data",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	// Check if user exists and is active
	var user models.User
	if err := config.DB.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Don't reveal if email exists for security
			c.JSON(http.StatusOK, models.SuccessResponse(
				"If the email exists, an OTP code has been sent",
				nil,
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Database error",
			"DATABASE_ERROR",
			"Please try again later",
		))
		return
	}

	// Generate 6-digit OTP
	otpCode, err := utils.GenerateOTP()
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to generate OTP",
			"OTP_GENERATION_ERROR",
			"Please try again later",
		))
		return
	}

	// Invalidate any existing unused OTPs for this email
	config.DB.Model(&models.OTP{}).
		Where("email = ? AND used = ?", req.Email, false).
		Update("used", true)

	// Create new OTP record (expires in 10 minutes)
	otp := models.OTP{
		Email:     req.Email,
		Code:      otpCode,
		Used:      false,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := config.DB.Create(&otp).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to save OTP",
			"DATABASE_ERROR",
			"Please try again later",
		))
		return
	}

	// Send OTP via email
	emailSubject := "Your Login OTP Code"
	emailBody := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
					<h2 style="color: #4CAF50;">Your OTP Code</h2>
					<p>Hello,</p>
					<p>Your OTP code for login is:</p>
					<div style="background-color: #f4f4f4; padding: 20px; text-align: center; margin: 20px 0; border-radius: 5px;">
						<h1 style="color: #4CAF50; margin: 0; font-size: 32px; letter-spacing: 5px;">%s</h1>
					</div>
					<p>This code will expire in 10 minutes.</p>
					<p>If you didn't request this code, please ignore this email.</p>
					<hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
					<p style="color: #666; font-size: 12px;">This is an automated email, please do not reply.</p>
				</div>
			</body>
		</html>
	`, otpCode)

	if err := utils.SendEmail(req.Email, emailSubject, emailBody); err != nil {
		// Log error but don't reveal it to user for security
		c.JSON(http.StatusOK, models.SuccessResponse(
			"If the email exists, an OTP code has been sent",
			nil,
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"OTP code has been sent to your email",
		nil,
	))
}

// VerifyOTP godoc
// @Summary Verify OTP and login
// @Description Verifies OTP code and returns JWT token for authentication
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.VerifyOTPRequest true "Email and OTP code"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/verify-otp [post]
func VerifyOTP(c *gin.Context) {
	var req models.VerifyOTPRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request data",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	// Find valid OTP
	var otp models.OTP
	if err := config.DB.Where("email = ? AND code = ? AND used = ?", req.Email, req.Code, false).
		First(&otp).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"Invalid or expired OTP code",
				"INVALID_OTP",
				"Please request a new OTP code",
			))
			return
		}
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Database error",
			"DATABASE_ERROR",
			"Please try again later",
		))
		return
	}

	// Check if OTP has expired
	if time.Now().After(otp.ExpiresAt) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"OTP code has expired",
			"OTP_EXPIRED",
			"Please request a new OTP code",
		))
		return
	}

	// Get user
	var user models.User
	if err := config.DB.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not found or inactive",
			"USER_NOT_FOUND",
			"Please register first",
		))
		return
	}

	// Mark OTP as used
	otp.Used = true
	config.DB.Save(&otp)

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := config.DB.Save(&user).Error; err != nil {
		// Log error but don't fail the login
		log.Printf("Failed to update last login time: %v", err)
	}

	// Generate JWT token
	token, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to generate token",
			"TOKEN_GENERATION_ERROR",
			"Please try again later",
		))
		return
	}

	userData := gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"name":          user.Name,
		"role":          user.Role,
		"last_login_at": user.LastLoginAt,
	}

	responseData := gin.H{
		"token": token,
		"user":  userData,
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Login successful",
		responseData,
	))
}
