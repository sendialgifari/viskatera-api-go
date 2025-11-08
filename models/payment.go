package models

import (
	"time"

	"gorm.io/gorm"
)

type Payment struct {
	ID            uint           `json:"id" gorm:"primaryKey"`
	UserID        uint           `json:"user_id" gorm:"not null;index:idx_payment_user_status,idx_payment_user_created"`
	User          User           `json:"user" gorm:"foreignKey:UserID"`
	PurchaseID    uint           `json:"purchase_id" gorm:"not null;index:idx_payment_purchase"`
	Purchase      VisaPurchase   `json:"purchase" gorm:"foreignKey:PurchaseID"`
	PaymentMethod string         `json:"payment_method" gorm:"not null;index:idx_payment_method"`
	Amount        float64        `json:"amount" gorm:"not null;index:idx_payment_amount"`
	Status        string         `json:"status" gorm:"default:'pending';index:idx_payment_user_status,idx_payment_status_created"` // pending, paid, expired, failed
	XenditID      string         `json:"xendit_id" gorm:"index:idx_payment_xendit,unique"`
	PaymentURL    string         `json:"payment_url"`
	CreatedAt     time.Time      `json:"created_at" gorm:"index:idx_payment_user_created,idx_payment_status_created"`
	UpdatedAt     time.Time      `json:"updated_at"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}
