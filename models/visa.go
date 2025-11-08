package models

import (
	"time"

	"gorm.io/gorm"
)

type Visa struct {
	ID              uint           `json:"id" gorm:"primaryKey"`
	Country         string         `json:"country" gorm:"not null;index:idx_visa_country_active"`
	Type            string         `json:"type" gorm:"not null;index:idx_visa_type_active"`
	Description     string         `json:"description"`
	Price           float64        `json:"price" gorm:"not null;index:idx_visa_price"`
	Duration        int            `json:"duration" gorm:"not null"` // in days
	VisaDocumentURL string         `json:"visa_document_url"`
	IsActive        bool           `json:"is_active" gorm:"default:true;index:idx_visa_country_active,idx_visa_type_active"`
	CreatedAt       time.Time      `json:"created_at" gorm:"index:idx_visa_created"`
	UpdatedAt       time.Time      `json:"updated_at"`
	DeletedAt       gorm.DeletedAt `json:"-" gorm:"index"`
}

type VisaOption struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	VisaID      uint           `json:"visa_id" gorm:"not null;index:idx_visa_option_visa_active"`
	Visa        Visa           `json:"visa" gorm:"foreignKey:VisaID"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"not null;index:idx_visa_option_price"`
	IsActive    bool           `json:"is_active" gorm:"default:true;index:idx_visa_option_visa_active"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

type VisaPurchase struct {
	ID           uint           `json:"id" gorm:"primaryKey"`
	UserID       uint           `json:"user_id" gorm:"not null;index:idx_purchase_user_status,idx_purchase_user_created"`
	User         User           `json:"user" gorm:"foreignKey:UserID"`
	VisaID       uint           `json:"visa_id" gorm:"not null;index:idx_purchase_visa"`
	Visa         Visa           `json:"visa" gorm:"foreignKey:VisaID"`
	VisaOptionID *uint          `json:"visa_option_id" gorm:"index:idx_purchase_option"`
	VisaOption   *VisaOption    `json:"visa_option" gorm:"foreignKey:VisaOptionID"`
	TotalPrice   float64        `json:"total_price" gorm:"not null;index:idx_purchase_price"`
	Status       string         `json:"status" gorm:"default:'pending';index:idx_purchase_user_status,idx_purchase_status_created"` // pending, completed, cancelled
	CreatedAt    time.Time      `json:"created_at" gorm:"index:idx_purchase_user_created,idx_purchase_status_created"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type PurchaseRequest struct {
	VisaID       uint   `json:"visa_id" binding:"required"`
	VisaOptionID *uint  `json:"visa_option_id"`
	Status       string `json:"status,omitempty"`
}
