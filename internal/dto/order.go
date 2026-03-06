package dto

import (
	"net/url"
	"time"
)

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
	UserID          uint                `json:"user_id"`
	UserName        string              `json:"user_name,omitempty"`
	UserEmail       string              `json:"user_email,omitempty"`
	ItemCount       int                 `json:"item_count"`
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

type AdminOrderListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=15" binding:"min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=pending confirmed processing shipping delivered cancelled"`
	FromDate string `form:"from_date" binding:"omitempty"`
	ToDate   string `form:"to_date" binding:"omitempty"`
	SortBy   string `form:"sort_by,default=created_at" binding:"omitempty,oneof=created_at total_amount status"`
	SortDir  string `form:"sort_dir,default=desc" binding:"omitempty,oneof=asc desc"`
}

type AdminUpdateOrderStatusRequest struct {
	Status string `form:"status" binding:"required,oneof=pending confirmed processing shipping delivered cancelled"`
}

func (q AdminOrderListRequest) URLParams() string {
	params := url.Values{}
	if q.Status != "" {
		params.Set("status", q.Status)
	}
	if q.FromDate != "" {
		params.Set("from_date", q.FromDate)
	}
	if q.ToDate != "" {
		params.Set("to_date", q.ToDate)
	}
	if q.SortBy != "" {
		params.Set("sort_by", q.SortBy)
	}
	if q.SortDir != "" {
		params.Set("sort_dir", q.SortDir)
	}
	return params.Encode()
}
