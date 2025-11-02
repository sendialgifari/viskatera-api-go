package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null"`
	User          User           `json:"user" gorm:"foreignKey:UserID"`
	PurchaseID    uint           `json:"purchase_id" gorm:"not null"`
	Purchase      VisaPurchase   `json:"purchase" gorm:"foreignKey:PurchaseID"`
	PaymentMethod string         `json:"payment_method" gorm:"not null"`
	Amount        float64        `json:"amount" gorm:"not null"`
	Status        string         `json:"status" gorm:"default:'pending'"` // pending, paid, expired, failed
	XenditID      string         `json:"xendit_id"`
	PaymentURL    string         `json:"payment_url"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

