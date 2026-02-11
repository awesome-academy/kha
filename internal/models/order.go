package models

import (
	"time"
)

type Order struct {
	ID              uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID          uint      `gorm:"not null;index" json:"user_id"`
	OrderNumber     string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"order_number"`
	TotalAmount     float64   `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Status          string    `gorm:"type:varchar(50);not null;default:pending;index" json:"status"`
	ShippingAddress string    `gorm:"type:text;not null" json:"shipping_address"`
	ShippingPhone   string    `gorm:"type:varchar(20);not null" json:"shipping_phone"`
	Notes           *string   `gorm:"type:text" json:"notes,omitempty"`
	CreatedAt       time.Time `gorm:"autoCreateTime;index" json:"created_at"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime" json:"updated_at"`

	// Relationships
	User          User                `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Items         []OrderItem         `gorm:"foreignKey:OrderID" json:"items,omitempty"`
	Ratings       []Rating            `gorm:"foreignKey:OrderID" json:"ratings,omitempty"`
	Notifications []OrderNotification `gorm:"foreignKey:OrderID" json:"notifications,omitempty"`
}

func (Order) TableName() string {
	return "orders"
}

// Status constants
const (
	OrderStatusPending    = "pending"
	OrderStatusConfirmed  = "confirmed"
	OrderStatusProcessing = "processing"
	OrderStatusShipping   = "shipping"
	OrderStatusDelivered  = "delivered"
	OrderStatusCancelled  = "cancelled"
)

type OrderItem struct {
	ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	OrderID      uint      `gorm:"not null;index" json:"order_id"`
	ProductID    uint      `gorm:"not null;index" json:"product_id"`
	ProductName  string    `gorm:"type:varchar(255);not null" json:"product_name"`
	ProductPrice float64   `gorm:"type:decimal(10,2);not null" json:"product_price"`
	Quantity     int       `gorm:"not null" json:"quantity"`
	Subtotal     float64   `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`

	// Relationships
	Order   Order   `gorm:"foreignKey:OrderID" json:"order,omitempty"`
	Product Product `gorm:"foreignKey:ProductID" json:"product,omitempty"`
}

func (OrderItem) TableName() string {
	return "order_items"
}
