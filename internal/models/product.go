package models

import (
	"time"

	"gorm.io/gorm"
)

type Product struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`
	CategoryID    uint           `gorm:"not null;index" json:"category_id"`
	Name          string         `gorm:"type:varchar(255);not null" json:"name"`
	Slug          string         `gorm:"type:varchar(255);uniqueIndex;not null" json:"slug"`
	Description   *string        `gorm:"type:text" json:"description,omitempty"`
	Classify      string         `gorm:"type:varchar(50);not null;index" json:"classify"`
	Price         float64        `gorm:"type:decimal(10,2);not null;index" json:"price"`
	Stock         int            `gorm:"not null;default:0" json:"stock"`
	RatingAverage float64        `gorm:"type:decimal(3,2);not null;default:0.00;index" json:"rating_average"`
	RatingCount   int            `gorm:"not null;default:0" json:"rating_count"`
	Status        string         `gorm:"type:varchar(50);not null;default:active;index" json:"status"`
	CreatedAt     time.Time      `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime" json:"updated_at"`
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`

	// Relationships
	Category   Category       `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
	Images     []ProductImage `gorm:"foreignKey:ProductID" json:"images,omitempty"`
	CartItems  []CartItem     `gorm:"foreignKey:ProductID" json:"cart_items,omitempty"`
	OrderItems []OrderItem    `gorm:"foreignKey:ProductID" json:"order_items,omitempty"`
	Ratings    []Rating       `gorm:"foreignKey:ProductID" json:"ratings,omitempty"`
}

func (Product) TableName() string {
	return "products"
}

// Classify constants
const (
	ClassifyFood  = "food"
	ClassifyDrink = "drink"
)

// Status constants
const (
	ProductStatusActive     = "active"
	ProductStatusInactive   = "inactive"
	ProductStatusOutOfStock = "out_of_stock"
)
