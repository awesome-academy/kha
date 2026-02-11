package models

import (
	"time"
)

type SocialAuth struct {
	ID             uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID         uint      `gorm:"not null;index" json:"user_id"`
	Provider       string    `gorm:"type:varchar(50);not null" json:"provider"`
	ProviderUserID string    `gorm:"type:varchar(255);not null" json:"provider_user_id"`
	AccessToken    *string   `gorm:"type:text" json:"-"`
	RefreshToken   *string   `gorm:"type:text" json:"-"`
	CreatedAt      time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt      time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user,omitempty"`
}

func (SocialAuth) TableName() string {
	return "social_auths"
}

// Provider constants
const (
	ProviderFacebook = "facebook"
	ProviderTwitter  = "twitter"
	ProviderGoogle   = "google"
)
