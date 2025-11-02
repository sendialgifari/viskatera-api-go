package models

import (
	"time"

	"gorm.io/gorm"
)

// ActivityAction represents the type of action performed
type ActivityAction string

const (
	ActionCreate ActivityAction = "create"
	ActionUpdate ActivityAction = "update"
	ActionDelete ActivityAction = "delete"
)

// ActivityEntity represents the entity type being acted upon
type ActivityEntity string

const (
	EntityUser     ActivityEntity = "user"
	EntityVisa     ActivityEntity = "visa"
	EntityPurchase ActivityEntity = "purchase"
	EntityPayment  ActivityEntity = "payment"
)

// ActivityLog represents an audit log entry
type ActivityLog struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	UserID      uint           `json:"user_id" gorm:"not null;index:idx_activity_user;index:idx_activity_user_created"`
	User        User           `json:"user" gorm:"foreignKey:UserID"`
	Action      ActivityAction `json:"action" gorm:"type:varchar(20);not null;index:idx_activity_action"`
	EntityType  ActivityEntity `json:"entity_type" gorm:"type:varchar(50);not null;index:idx_activity_entity_type;index:idx_activity_entity"`
	EntityID    uint           `json:"entity_id" gorm:"not null;index:idx_activity_entity_id;index:idx_activity_entity"`
	EntityName  string         `json:"entity_name" gorm:"size:255"`
	Description string         `json:"description" gorm:"type:text"`
	Changes     string         `json:"changes" gorm:"type:jsonb"`
	IPAddress   string         `json:"ip_address" gorm:"size:45"`
	UserAgent   string         `json:"user_agent" gorm:"type:text"`
	CreatedAt   time.Time      `json:"created_at" gorm:"index:idx_activity_created;index:idx_activity_user_created"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// TableName specifies the table name for ActivityLog
func (ActivityLog) TableName() string {
	return "activity_logs"
}
