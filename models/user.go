package models

import (
	"time"

	"gorm.io/gorm"
)

type UserRole string

const (
	RoleCustomer UserRole = "customer"
	RoleAdmin    UserRole = "admin"
)

type User struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Email       string         `json:"email" gorm:"unique;not null;index:idx_user_email_active"`
	Password    string         `json:"-" gorm:"not null"`
	Name        string         `json:"name" gorm:"not null;index:idx_user_name"`
	AvatarURL   string         `json:"avatar_url"`
	GoogleID    string         `json:"google_id" gorm:"index:idx_user_google,unique"`
	Role        UserRole       `json:"role" gorm:"type:varchar(20);default:'customer';index:idx_user_role_active"`
	IsActive    bool           `json:"is_active" gorm:"default:true;index:idx_user_email_active,idx_user_role_active"`
	LastLoginAt *time.Time     `json:"last_login_at" gorm:"index:idx_user_last_login"`
	CreatedAt   time.Time      `json:"created_at" gorm:"index:idx_user_created"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
	Role     string `json:"role,omitempty" binding:"omitempty,oneof=customer admin"`
}
