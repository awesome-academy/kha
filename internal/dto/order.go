package dto

import "time"

type CreateOrderRequest struct {
	ShippingAddress string  `json:"shipping_address" binding:"required,max=2000"`
	ShippingPhone   string  `json:"shipping_phone" binding:"required,min=8,max=20"`
	Notes           *string `json:"notes" binding:"omitempty,max=5000"`
}

type OrderItemResponse struct {
	ID           uint    `json:"id"`
	ProductID    uint    `json:"product_id"`
	ProductName  string  `json:"product_name"`
	ProductPrice float64 `json:"product_price"`
	Quantity     int     `json:"quantity"`
	Subtotal     float64 `json:"subtotal"`
}

type OrderResponse struct {
	ID              uint                `json:"id"`
	OrderNumber     string              `json:"order_number"`
	TotalAmount     float64             `json:"total_amount"`
	Status          string              `json:"status"`
	ShippingAddress string              `json:"shipping_address"`
	ShippingPhone   string              `json:"shipping_phone"`
	Notes           *string             `json:"notes,omitempty"`
	Items           []OrderItemResponse `json:"items,omitempty"`
	CreatedAt       time.Time           `json:"created_at"`
	UpdatedAt       time.Time           `json:"updated_at"`
}

type OrderListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=20" binding:"min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=pending confirmed processing shipping delivered cancelled"`
	FromDate string `form:"from_date" binding:"omitempty"`
	ToDate   string `form:"to_date" binding:"omitempty"`
}
