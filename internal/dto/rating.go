package dto

import "time"

type CreateRatingRequest struct {
	Rating  uint8   `json:"rating" binding:"required,min=1,max=5"`
	Comment *string `json:"comment" binding:"omitempty,max=2000"`
}

type UpdateRatingRequest struct {
	Rating  uint8   `json:"rating" binding:"required,min=1,max=5"`
	Comment *string `json:"comment" binding:"omitempty,max=2000"`
}

type RatingResponse struct {
	ID         uint      `json:"id"`
	UserID     uint      `json:"user_id"`
	UserName   string    `json:"user_name"`
	UserAvatar *string   `json:"user_avatar,omitempty"`
	ProductID  uint      `json:"product_id"`
	OrderID    *uint     `json:"order_id,omitempty"`
	Rating     uint8     `json:"rating"`
	Comment    *string   `json:"comment,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type RatingListRequest struct {
	Page     int `form:"page,default=1" binding:"min=1"`
	PageSize int `form:"page_size,default=20" binding:"min=1,max=100"`
}
