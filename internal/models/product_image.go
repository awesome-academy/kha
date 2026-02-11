package models

import (
	"time"
)

type ProductImage struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	ProductID uint      `gorm:"not null;index" json:"product_id"`
	ImageURL  string    `gorm:"type:varchar(500);not null" json:"image_url"`
	AltText   *string   `gorm:"type:varchar(255)" json:"alt_text,omitempty"`
	SortOrder int       `gorm:"not null;default:0;index" json:"sort_order"`
	IsPrimary bool      `gorm:"not null;default:false" json:"is_primary"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (ProductImage) TableName() string {
	return "product_images"
}
