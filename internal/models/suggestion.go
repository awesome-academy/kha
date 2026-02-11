package models

import (
	"time"
)

type Suggestion struct {
	ID          uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID      uint      `gorm:"not null;index" json:"user_id"`
	Name        string    `gorm:"type:varchar(255);not null" json:"name"`
	Description *string   `gorm:"type:text" json:"description,omitempty"`
	Classify    string    `gorm:"type:varchar(50);not null" json:"classify"`
	CategoryID  *uint     `gorm:"index" json:"category_id,omitempty"`
	Status      string    `gorm:"type:varchar(50);not null;default:pending;index" json:"status"`
	AdminNote   *string   `gorm:"type:text" json:"admin_note,omitempty"`
	CreatedAt   time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User     User      `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Category *Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

func (Suggestion) TableName() string {
	return "suggestions"
}

// Status constants
const (
	SuggestionStatusPending  = "pending"
	SuggestionStatusApproved = "approved"
	SuggestionStatusRejected = "rejected"
)
