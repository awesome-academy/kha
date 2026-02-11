package models

import (
	"time"

	"gorm.io/gorm"
)

type Category struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug        string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description *string        `gorm:"type:text" json:"description,omitempty"`
	ImageURL    *string        `gorm:"type:varchar(500)" json:"image_url,omitempty"`
	SortOrder   int            `gorm:"not null;default:0;index" json:"sort_order"`
	Status      string         `gorm:"type:varchar(50);not null;default:active;index" json:"status"`
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Products    []Product    `gorm:"foreignKey:CategoryID" json:"products,omitempty"`
	Suggestions []Suggestion `gorm:"foreignKey:CategoryID" json:"suggestions,omitempty"`
}

func (Category) TableName() string {
	return "categories"
}

// Status constants
const (
	CategoryStatusActive   = "active"
	CategoryStatusInactive = "inactive"
)
