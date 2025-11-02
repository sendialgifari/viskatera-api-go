package controllers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
	"viskatera-api-go/config"
	"viskatera-api-go/models"
	"viskatera-api-go/utils"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email, password, name and optional role
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.RegisterRequest true "User registration data"
// @Success 201 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 409 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /register [post]
func Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request data",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	// Check if user already exists
	var existingUser models.User
	if err := config.DB.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		c.JSON(http.StatusConflict, models.ErrorResponse(
			"User with this email already exists",
			"USER_EXISTS",
			"Please use a different email address",
		))
		return
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to process registration",
			"PASSWORD_HASH_ERROR",
			"Please try again later",
		))
		return
	}

	// Set default role
	role := models.RoleCustomer
	if req.Role == "admin" {
		role = models.RoleAdmin
	}

	// Create user
	user := models.User{
		Email:    req.Email,
		Password: hashedPassword,
		Name:     req.Name,
		Role:     role,
	}

	if err := config.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create user",
			"USER_CREATION_ERROR",
			"Please try again later",
		))
		return
	}

	userData := gin.H{
		"id":    user.ID,
		"email": user.Email,
		"name":  user.Name,
		"role":  user.Role,
	}

	c.JSON(http.StatusCreated, models.SuccessResponse(
		"User registered successfully",
		gin.H{"user": userData},
	))
}

// Login godoc
// @Summary Login user
// @Description Login user with email and password
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "User login credentials"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /login [post]
func Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid request data",
			"VALIDATION_ERROR",
			err.Error(),
		))
		return
	}

	// Find user by email
	var user models.User
	if err := config.DB.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse(
				"Invalid email or password",
				"INVALID_CREDENTIALS",
				"Please check your email and password",
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

	// Check password
	if !utils.CheckPasswordHash(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"Invalid email or password",
			"INVALID_CREDENTIALS",
			"Please check your email and password",
		))
		return
	}

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

// ForgotPassword godoc
// @Summary Request password reset
// @Description Sends password reset link to user's email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.ForgotPasswordRequest true "Email address"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/forgot-password [post]
func ForgotPassword(c *gin.Context) {
	var req models.ForgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request data", "VALIDATION_ERROR", err.Error()))
		return
	}

	var user models.User
	if err := config.DB.Where("email = ? AND is_active = ?", req.Email, true).First(&user).Error; err != nil {
		// Do not reveal if email exists
		c.JSON(http.StatusOK, models.SuccessResponse("If the email exists, a reset link has been sent", nil))
		return
	}

	token, err := utils.GenerateSecureToken(32)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to create token", "TOKEN_ERROR", err.Error()))
		return
	}

	prt := models.PasswordResetToken{
		UserID:    user.ID,
		Token:     token,
		ExpiresAt: time.Now().Add(30 * time.Minute),
	}
	if err := config.DB.Create(&prt).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to save token", "DATABASE_ERROR", err.Error()))
		return
	}

	resetURL := fmt.Sprintf("%s/reset-password?token=%s", os.Getenv("APP_BASE_URL"), token)
	emailSubject := "Password Reset Request"
	emailBody := fmt.Sprintf(`
		<html>
			<body style="font-family: Arial, sans-serif; line-height: 1.6; color: #333;">
				<div style="max-width: 600px; margin: 0 auto; padding: 20px;">
					<h2 style="color: #4CAF50;">Password Reset Request</h2>
					<p>Hello,</p>
					<p>You have requested to reset your password. Click the link below to reset it:</p>
					<div style="text-align: center; margin: 30px 0;">
						<a href="%s" style="background-color: #4CAF50; color: white; padding: 12px 30px; text-decoration: none; border-radius: 5px; display: inline-block;">Reset Password</a>
					</div>
					<p>Or copy and paste this link into your browser:</p>
					<p style="word-break: break-all; color: #666;">%s</p>
					<p>This link will expire in 30 minutes.</p>
					<p>If you didn't request a password reset, please ignore this email.</p>
					<hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
					<p style="color: #666; font-size: 12px;">This is an automated email, please do not reply.</p>
				</div>
			</body>
		</html>
	`, resetURL, resetURL)
	_ = utils.SendEmail(user.Email, emailSubject, emailBody)

	c.JSON(http.StatusOK, models.SuccessResponse("If the email exists, a reset link has been sent", nil))
}

// ResetPassword godoc
// @Summary Reset password with token
// @Description Resets user password using reset token from email
// @Tags Authentication
// @Accept json
// @Produce json
// @Param request body models.ResetPasswordRequest true "Reset token and new password"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/reset-password [post]
func ResetPassword(c *gin.Context) {
	var req models.ResetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request data", "VALIDATION_ERROR", err.Error()))
		return
	}

	var prt models.PasswordResetToken
	if err := config.DB.Where("token = ? AND used = ?", req.Token, false).First(&prt).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid or expired token", "INVALID_TOKEN", ""))
		return
	}
	if time.Now().After(prt.ExpiresAt) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Token expired", "TOKEN_EXPIRED", ""))
		return
	}

	var user models.User
	if err := config.DB.First(&user, prt.UserID).Error; err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("User not found", "USER_NOT_FOUND", ""))
		return
	}

	hashed, err := utils.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to hash password", "HASH_ERROR", ""))
		return
	}

	user.Password = hashed
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to update password", "DATABASE_ERROR", ""))
		return
	}

	prt.Used = true
	config.DB.Save(&prt)

	c.JSON(http.StatusOK, models.SuccessResponse("Password updated successfully", nil))
}

// UpdateUser godoc
// @Summary Update user profile
// @Description Update user profile information and/or password
// @Tags Authentication
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.UpdateUserRequest true "User update data"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 404 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /user [put]
func UpdateUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse("Unauthorized", "UNAUTHORIZED", ""))
		return
	}

	var req models.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid request data", "VALIDATION_ERROR", err.Error()))
		return
	}

	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse("User not found", "USER_NOT_FOUND", ""))
		return
	}

	if req.Name != "" {
		user.Name = req.Name
	}
	if req.Email != "" {
		user.Email = req.Email
	}

	if req.NewPassword != "" {
		// require current password if user has a password (not Google-only)
		if user.Password != "" {
			if req.CurrentPassword == "" || !utils.CheckPasswordHash(req.CurrentPassword, user.Password) {
				c.JSON(http.StatusBadRequest, models.ErrorResponse("Current password is incorrect", "INVALID_PASSWORD", ""))
				return
			}
		}
		hashed, err := utils.HashPassword(req.NewPassword)
		if err != nil {
			c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to hash password", "HASH_ERROR", ""))
			return
		}
		user.Password = hashed
	}

	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to update user", "DATABASE_ERROR", ""))
		return
	}
	c.JSON(http.StatusOK, models.SuccessResponse("User updated successfully", gin.H{"user": gin.H{"id": user.ID, "email": user.Email, "name": user.Name, "avatar_url": user.AvatarURL}}))
}

// GoogleLogin godoc
// @Summary Login with Google OAuth
// @Description Redirects to Google OAuth consent screen
// @Tags Authentication
// @Success 302
// @Router /auth/google/login [get]
func GoogleLogin(c *gin.Context) {
	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	redirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	scope := "https://www.googleapis.com/auth/userinfo.email https://www.googleapis.com/auth/userinfo.profile"
	state, _ := utils.GenerateSecureToken(16)
	// Store state in cookie for CSRF protection
	http.SetCookie(c.Writer, &http.Cookie{Name: "oauth_state", Value: state, Path: "/", Expires: time.Now().Add(10 * time.Minute), HttpOnly: true, SameSite: http.SameSiteLaxMode})
	authURL := fmt.Sprintf("https://accounts.google.com/o/oauth2/v2/auth?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s&access_type=offline", clientID, redirectURL, url.QueryEscape(scope), state)
	c.Redirect(http.StatusFound, authURL)
}

// GoogleCallback godoc
// @Summary Google OAuth callback
// @Description Handles Google OAuth callback and returns JWT token
// @Tags Authentication
// @Param code query string true "OAuth code"
// @Param state query string true "OAuth state"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /auth/google/callback [get]
func GoogleCallback(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")
	// Validate state
	if cookie, err := c.Request.Cookie("oauth_state"); err != nil || cookie.Value != state {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Invalid state", "INVALID_STATE", ""))
		return
	}

	tokenResp, err := exchangeCodeForToken(code)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("OAuth exchange failed", "OAUTH_ERROR", err.Error()))
		return
	}

	info, err := fetchGoogleUserInfo(tokenResp.AccessToken)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse("Failed to fetch user info", "OAUTH_ERROR", err.Error()))
		return
	}

	// Upsert user by google id or email
	var user models.User
	if err := config.DB.Where("google_id = ? OR email = ?", info.ID, info.Email).First(&user).Error; err != nil {
		user = models.User{Email: info.Email, Name: info.Name, GoogleID: info.ID, Password: utils.MustHashPlaceholder()}
		_ = config.DB.Create(&user).Error
	} else {
		if user.GoogleID == "" {
			user.GoogleID = info.ID
			config.DB.Save(&user)
		}
	}

	// Update last login time
	now := time.Now()
	user.LastLoginAt = &now
	if err := config.DB.Save(&user).Error; err != nil {
		// Log error but don't fail the login
		log.Printf("Failed to update last login time: %v", err)
	}

	jwt, err := utils.GenerateJWT(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse("Failed to generate token", "TOKEN_ERROR", ""))
		return
	}

	userData := gin.H{
		"id":            user.ID,
		"email":         user.Email,
		"name":          user.Name,
		"role":          user.Role,
		"last_login_at": user.LastLoginAt,
	}

	c.JSON(http.StatusOK, models.SuccessResponse("Login successful", gin.H{
		"token": jwt,
		"user":  userData,
	}))
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

func exchangeCodeForToken(code string) (tokenResponse, error) {
	form := url.Values{}
	form.Set("code", code)
	form.Set("client_id", os.Getenv("GOOGLE_CLIENT_ID"))
	form.Set("client_secret", os.Getenv("GOOGLE_CLIENT_SECRET"))
	form.Set("redirect_uri", os.Getenv("GOOGLE_REDIRECT_URL"))
	form.Set("grant_type", "authorization_code")
	req, _ := http.NewRequestWithContext(context.Background(), "POST", "https://oauth2.googleapis.com/token", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return tokenResponse{}, err
	}
	defer res.Body.Close()
	var tr tokenResponse
	if err := json.NewDecoder(res.Body).Decode(&tr); err != nil {
		return tokenResponse{}, err
	}
	if tr.AccessToken == "" {
		return tokenResponse{}, fmt.Errorf("no access token")
	}
	return tr, nil
}

type googleUser struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

func fetchGoogleUserInfo(accessToken string) (googleUser, error) {
	req, _ := http.NewRequest("GET", "https://www.googleapis.com/oauth2/v2/userinfo", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return googleUser{}, err
	}
	defer res.Body.Close()
	var info googleUser
	if err := json.NewDecoder(res.Body).Decode(&info); err != nil {
		return googleUser{}, err
	}
	if info.Email == "" {
		return googleUser{}, fmt.Errorf("no email")
	}
	return info, nil
}
