package dto

import (
	"net/url"
	"time"
)

type CreateSuggestionRequest struct {
	Name        string  `json:"name" binding:"required,min=2,max=255"`
	Description *string `json:"description" binding:"omitempty,max=5000"`
	Classify    string  `json:"classify" binding:"required,oneof=food drink"`
	CategoryID  *uint   `json:"category_id" binding:"omitempty"`
}

type SuggestionResponse struct {
	ID           uint      `json:"id"`
	UserID       uint      `json:"user_id"`
	UserName     string    `json:"user_name,omitempty"`
	Name         string    `json:"name"`
	Description  *string   `json:"description,omitempty"`
	Classify     string    `json:"classify"`
	CategoryID   *uint     `json:"category_id,omitempty"`
	CategoryName string    `json:"category_name,omitempty"`
	Status       string    `json:"status"`
	AdminNote    *string   `json:"admin_note,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type AdminSuggestionListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=15" binding:"min=1,max=100"`
	Status   string `form:"status" binding:"omitempty,oneof=pending approved rejected"`
	Classify string `form:"classify" binding:"omitempty,oneof=food drink"`
	Search   string `form:"search" binding:"omitempty,max=255"`
	SortBy   string `form:"sort_by,default=created_at" binding:"omitempty,oneof=created_at status name"`
	SortDir  string `form:"sort_dir,default=desc" binding:"omitempty,oneof=asc desc"`
}

type AdminUpdateSuggestionStatusRequest struct {
	Status    string  `form:"status" binding:"required,oneof=approved rejected"`
	AdminNote *string `form:"admin_note" binding:"omitempty,max=5000"`
}

func (q AdminSuggestionListRequest) URLParams() string {
	params := url.Values{}
	if q.Status != "" {
		params.Set("status", q.Status)
	}
	if q.Classify != "" {
		params.Set("classify", q.Classify)
	}
	if q.Search != "" {
		params.Set("search", q.Search)
	}
	if q.SortBy != "" {
		params.Set("sort_by", q.SortBy)
	}
	if q.SortDir != "" {
		params.Set("sort_dir", q.SortDir)
	}
	return params.Encode()
}
