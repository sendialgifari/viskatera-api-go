package models

import (
	"time"

	"gorm.io/gorm"
)

type Visa struct {
	ID                uint           `json:"id" gorm:"primaryKey"`
	Country           string         `json:"country" gorm:"not null"`
	Type              string         `json:"type" gorm:"not null"`
	Description       string         `json:"description"`
	Price             float64        `json:"price" gorm:"not null"`
	Duration          int            `json:"duration" gorm:"not null"` // in days
	VisaDocumentURL   string         `json:"visa_document_url"`
	IsActive          bool           `json:"is_active" gorm:"default:true"`
	CreatedAt         time.Time      `json:"created_at"`
	UpdatedAt         time.Time      `json:"updated_at"`
	DeletedAt         gorm.DeletedAt `json:"-" gorm:"index"`
}

type VisaOption struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	VisaID      uint           `json:"visa_id" gorm:"not null"`
	Visa        Visa           `json:"visa" gorm:"foreignKey:VisaID"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"not null"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type VisaPurchase struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id" gorm:"not null"`
	User         User           `json:"user" gorm:"foreignKey:UserID"`
	VisaID       uint           `json:"visa_id" gorm:"not null"`
	Visa         Visa           `json:"visa" gorm:"foreignKey:VisaID"`
	VisaOptionID *uint          `json:"visa_option_id"`
	VisaOption   *VisaOption    `json:"visa_option" gorm:"foreignKey:VisaOptionID"`
	TotalPrice   float64        `json:"total_price" gorm:"not null"`
	Status       string         `json:"status" gorm:"default:'pending'"` // pending, completed, cancelled
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type PurchaseRequest struct {
	VisaID       uint   `json:"visa_id" binding:"required"`
	VisaOptionID *uint  `json:"visa_option_id"`
	Status       string `json:"status,omitempty"`
}
