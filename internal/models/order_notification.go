package models

import (
	"time"
)

type OrderNotification struct {
	ID           uint       `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID      uint       `gorm:"not null;index" json:"order_id"`
	Type         string     `gorm:"type:varchar(50);not null;index" json:"type"`
	Status       string     `gorm:"type:varchar(50);not null;default:pending;index" json:"status"`
	Recipient    string     `gorm:"type:varchar(255);not null" json:"recipient"`
	Message      *string    `gorm:"type:text" json:"message,omitempty"`
	ErrorMessage *string    `gorm:"type:text" json:"error_message,omitempty"`
	SentAt       *time.Time `gorm:"type:timestamp" json:"sent_at,omitempty"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	Order *Order `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

func (OrderNotification) TableName() string {
	return "order_notifications"
}

// Type constants
const (
	NotificationTypeEmail    = "email"
	NotificationTypeChatwork = "chatwork"
)

// Status constants
const (
	NotificationStatusPending = "pending"
	NotificationStatusSent    = "sent"
	NotificationStatusFailed  = "failed"
)
