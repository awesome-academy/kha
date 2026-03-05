package dto

type CartItemResponse struct {
	ID        uint    `json:"id"`
	ProductID uint    `json:"product_id"`
	Name      string  `json:"name"`
	Price     float64 `json:"price"`
	Quantity  int     `json:"quantity"`
	Subtotal  float64 `json:"subtotal"`
	ImageURL  string  `json:"image_url,omitempty"`
}

type CartResponse struct {
	ID          uint               `json:"id"`
	Items       []CartItemResponse `json:"items"`
	TotalItems  int                `json:"total_items"`
	TotalAmount float64            `json:"total_amount"`
}

type AddCartItemRequest struct {
	ProductID uint `json:"product_id" binding:"required"`
	Quantity  int  `json:"quantity" binding:"required,min=1"`
}

type UpdateCartItemRequest struct {
	Quantity int `json:"quantity" binding:"required,min=1"`
}

type RemoveCartItemRequest struct {
	ProductID uint `uri:"product_id" binding:"required"`
}
