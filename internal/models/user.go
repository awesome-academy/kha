package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID              uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	Email           string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash    *string        `gorm:"type:varchar(255)" json:"-"`
	FullName        string         `gorm:"type:varchar(255);not null" json:"full_name"`
	Phone           *string        `gorm:"type:varchar(20)" json:"phone,omitempty"`
	Address         *string        `gorm:"type:text" json:"address,omitempty"`
	AvatarURL       *string        `gorm:"type:varchar(500)" json:"avatar_url,omitempty"`
	Role            string         `gorm:"type:varchar(50);not null;default:user;index" json:"role"`
	Status          string         `gorm:"type:varchar(50);not null;default:active;index" json:"status"`
	EmailVerifiedAt *time.Time     `gorm:"type:timestamp" json:"email_verified_at,omitempty"`
	CreatedAt       time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt       time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt       gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	SocialAuths []SocialAuth `gorm:"foreignKey:UserID" json:"social_auths,omitempty"`
	Cart        *Cart        `gorm:"foreignKey:UserID" json:"cart,omitempty"`
	Orders      []Order      `gorm:"foreignKey:UserID" json:"orders,omitempty"`
	Ratings     []Rating     `gorm:"foreignKey:UserID" json:"ratings,omitempty"`
	Suggestions []Suggestion `gorm:"foreignKey:UserID" json:"suggestions,omitempty"`
}

func (User) TableName() string {
	return "users"
}

// Role constants
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)

// Status constants
const (
	UserStatusActive   = "active"
	UserStatusInactive = "inactive"
	UserStatusBanned   = "banned"
)
