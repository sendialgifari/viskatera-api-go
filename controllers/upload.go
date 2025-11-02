package controllers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"
	"viskatera-api-go/config"
	"viskatera-api-go/models"

	"viskatera-api-go/utils"

	"github.com/gin-gonic/gin"
)

// UploadAvatar godoc
// @Summary Upload user avatar
// @Description Upload avatar image for user profile
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "Avatar image file"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /uploads/avatar [post]
func UploadAvatar(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to upload avatar",
		))
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid file upload",
			"FILE_ERROR",
			err.Error(),
		))
		return
	}

	// Validate file size
	maxSize := getEnvAsInt("MAX_UPLOAD_SIZE", 10485760) // 10MB default
	if file.Size > int64(maxSize) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"File too large",
			"FILE_TOO_LARGE",
			fmt.Sprintf("File size must be less than %d bytes", maxSize),
		))
		return
	}

	// Validate file extension
	ext := filepath.Ext(file.Filename)
	allowedExts := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid file type",
			"INVALID_FILE_TYPE",
			"Only image files are allowed (.jpg, .jpeg, .png, .gif, .webp)",
		))
		return
	}

	// Create upload directory if it doesn't exist
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	avatarDir := filepath.Join(uploadDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create upload directory",
			"DIRECTORY_ERROR",
			err.Error(),
		))
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("%d_%d%s", userID.(uint), time.Now().Unix(), ext)
	tempPath := filepath.Join(avatarDir, "temp_"+filename)
	finalPath := filepath.Join(avatarDir, filename)

	// Save file temporarily
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to save file",
			"SAVE_ERROR",
			err.Error(),
		))
		return
	}

	// Compress image if it's a supported image format
	imageExts := []string{".jpg", ".jpeg", ".png"}
	isImage := false
	for _, imageExt := range imageExts {
		if ext == imageExt {
			isImage = true
			break
		}
	}

	if isImage {
		// Compress the image
		if err := utils.CompressAndSave(tempPath, finalPath, utils.DefaultAvatarConfig()); err != nil {
			// If compression fails, use the original file
			os.Rename(tempPath, finalPath)
		} else {
			// Remove temp file after successful compression
			os.Remove(tempPath)
		}
	} else {
		// For non-image files, just move temp to final
		os.Rename(tempPath, finalPath)
	}

	// Get relative path
	relativePath := fmt.Sprintf("uploads/avatars/%s", filename)

	// Update user's avatar URL
	var user models.User
	if err := config.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to get user",
			"DATABASE_ERROR",
			err.Error(),
		))
		return
	}

	// Delete old avatar if exists
	if user.AvatarURL != "" {
		oldPath := user.AvatarURL
		if _, err := os.Stat(oldPath); err == nil {
			os.Remove(oldPath)
		}
	}

	// Update avatar URL
	user.AvatarURL = relativePath
	if err := config.DB.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to update user",
			"DATABASE_ERROR",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Avatar uploaded successfully",
		gin.H{
			"avatar_url": relativePath,
			"filename":   filename,
		},
	))
}

// UploadVisaDocument godoc
// @Summary Upload visa document
// @Description Upload visa document file
// @Tags Upload
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param visa_id path int true "Visa ID"
// @Param file formData file true "Document file"
// @Success 200 {object} models.APIResponse
// @Failure 400 {object} models.APIResponse
// @Failure 401 {object} models.APIResponse
// @Failure 500 {object} models.APIResponse
// @Router /uploads/visa/{visa_id} [post]
func UploadVisaDocument(c *gin.Context) {
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse(
			"User not authenticated",
			"UNAUTHORIZED",
			"Please login to upload visa document",
		))
		return
	}

	visaID, err := strconv.Atoi(c.Param("visa_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid visa ID",
			"INVALID_ID",
			"Visa ID must be a valid number",
		))
		return
	}

	// Get file from form
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid file upload",
			"FILE_ERROR",
			err.Error(),
		))
		return
	}

	// Validate file size
	maxSize := getEnvAsInt("MAX_UPLOAD_SIZE", 10485760) // 10MB default
	if file.Size > int64(maxSize) {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"File too large",
			"FILE_TOO_LARGE",
			fmt.Sprintf("File size must be less than %d bytes", maxSize),
		))
		return
	}

	// Validate file extension
	ext := filepath.Ext(file.Filename)
	allowedExts := []string{".pdf", ".jpg", ".jpeg", ".png", ".doc", ".docx"}
	isAllowed := false
	for _, allowedExt := range allowedExts {
		if ext == allowedExt {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid file type",
			"INVALID_FILE_TYPE",
			"Allowed file types: .pdf, .jpg, .jpeg, .png, .doc, .docx",
		))
		return
	}

	// Get visa
	var visa models.Visa
	if err := config.DB.First(&visa, visaID).Error; err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"Visa not found",
			"VISA_NOT_FOUND",
			"Visa with this ID does not exist",
		))
		return
	}

	// Create upload directory if it doesn't exist
	uploadDir := getEnv("UPLOAD_DIR", "./uploads")
	docDir := filepath.Join(uploadDir, "visas")
	if err := os.MkdirAll(docDir, 0755); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to create upload directory",
			"DIRECTORY_ERROR",
			err.Error(),
		))
		return
	}

	// Generate unique filename
	filename := fmt.Sprintf("visa_%d_%d%s", visaID, time.Now().Unix(), ext)
	tempPath := filepath.Join(docDir, "temp_"+filename)
	finalPath := filepath.Join(docDir, filename)

	// Save file temporarily
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to save file",
			"SAVE_ERROR",
			err.Error(),
		))
		return
	}

	// Compress image if it's a supported image format
	imageExts := []string{".jpg", ".jpeg", ".png"}
	isImage := false
	for _, imageExt := range imageExts {
		if ext == imageExt {
			isImage = true
			break
		}
	}

	if isImage {
		// Compress the image
		if err := utils.CompressAndSave(tempPath, finalPath, utils.DefaultVisaDocConfig()); err != nil {
			// If compression fails, use the original file
			os.Rename(tempPath, finalPath)
		} else {
			// Remove temp file after successful compression
			os.Remove(tempPath)
		}
	} else {
		// For non-image files like PDF, DOC, just move temp to final
		os.Rename(tempPath, finalPath)
	}

	// Get relative path
	relativePath := fmt.Sprintf("uploads/visas/%s", filename)

	// Delete old document if exists
	if visa.VisaDocumentURL != "" {
		oldPath := visa.VisaDocumentURL
		if _, err := os.Stat(oldPath); err == nil {
			os.Remove(oldPath)
		}
	}

	// Update visa document URL
	visa.VisaDocumentURL = relativePath
	if err := config.DB.Save(&visa).Error; err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to update visa",
			"DATABASE_ERROR",
			err.Error(),
		))
		return
	}

	c.JSON(http.StatusOK, models.SuccessResponse(
		"Visa document uploaded successfully",
		gin.H{
			"visa_document_url": relativePath,
			"filename":          filename,
		},
	))
}

// ServeUploadedFile serves uploaded files
// Note: This is a static file handler, not an API endpoint
// Files are served directly via Gin static file serving
func ServeUploadedFile(c *gin.Context) {
	path := c.Param("path")
	fullPath := filepath.Join(getEnv("UPLOAD_DIR", "./uploads"), path)

	// Security check - prevent directory traversal
	if !isPathSafe(fullPath, getEnv("UPLOAD_DIR", "./uploads")) {
		c.JSON(http.StatusForbidden, models.ErrorResponse(
			"Access denied",
			"ACCESS_DENIED",
			"Invalid file path",
		))
		return
	}

	// Check if file exists
	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		c.JSON(http.StatusNotFound, models.ErrorResponse(
			"File not found",
			"FILE_NOT_FOUND",
			"The requested file does not exist",
		))
		return
	}

	if fileInfo.IsDir() {
		c.JSON(http.StatusBadRequest, models.ErrorResponse(
			"Invalid file path",
			"INVALID_PATH",
			"Requested path is a directory",
		))
		return
	}

	// Open and serve file
	file, err := os.Open(fullPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse(
			"Failed to open file",
			"FILE_ERROR",
			err.Error(),
		))
		return
	}
	defer file.Close()

	// Serve the file
	c.File(fullPath)
}

// Helper functions
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func getEnvAsInt(key string, defaultValue int) int {
	valueStr := getEnv(key, "")
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value
}

func isPathSafe(requestedPath, baseDir string) bool {
	// Clean paths
	requestedPath = filepath.Clean(requestedPath)
	baseDir = filepath.Clean(baseDir)

	// Get absolute paths
	absRequested, _ := filepath.Abs(requestedPath)
	absBase, _ := filepath.Abs(baseDir)

	// Check if requested path is within base directory
	return len(absRequested) >= len(absBase) && absRequested[:len(absBase)] == absBase
}
