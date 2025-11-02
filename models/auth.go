package models

import (
	"time"

	"gorm.io/gorm"
)

type PasswordResetToken struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	UserID    uint           `json:"user_id" gorm:"not null;index"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	Token     string         `json:"token" gorm:"uniqueIndex;size:128;not null"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"index"`
	Used      bool           `json:"used" gorm:"default:false"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type UpdateUserRequest struct {
	Name            string `json:"name" binding:"omitempty,min=2"`
	Email           string `json:"email" binding:"omitempty,email"`
	CurrentPassword string `json:"current_password" binding:"omitempty,min=6"`
	NewPassword     string `json:"new_password" binding:"omitempty,min=6"`
}

// OTP Model for login authentication
type OTP struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Email     string         `json:"email" gorm:"not null;index"`
	Code      string         `json:"code" gorm:"not null;size:6;index"`
	Used      bool           `json:"used" gorm:"default:false"`
	ExpiresAt time.Time      `json:"expires_at" gorm:"index"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

// OTP Request Models
type RequestOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
}

type VerifyOTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	Code  string `json:"code" binding:"required,len=6"`
}
