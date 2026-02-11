package models

import (
	"time"
)

type Rating struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    uint      `gorm:"not null;uniqueIndex:uk_user_product" json:"user_id"`
	ProductID uint      `gorm:"not null;index;uniqueIndex:uk_user_product" json:"product_id"`
	OrderID   *uint     `gorm:"index" json:"order_id,omitempty"`
	Rating    uint8     `gorm:"not null" json:"rating"`
	Comment   *string   `gorm:"type:text" json:"comment,omitempty"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User    User    `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
	Order   *Order  `gorm:"foreignKey:OrderID" json:"order,omitempty"`
}

func (Rating) TableName() string {
	return "ratings"
}
