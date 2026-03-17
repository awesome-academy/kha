package dto

import "net/url"

type AdminUserListRequest struct {
	Page     int    `form:"page,default=1" binding:"min=1"`
	PageSize int    `form:"page_size,default=15" binding:"min=1,max=100"`
	Search   string `form:"search" binding:"omitempty,max=255"`
	Status   string `form:"status" binding:"omitempty,oneof=active inactive banned"`
	Role     string `form:"role" binding:"omitempty,oneof=user admin"`
	SortBy   string `form:"sort_by,default=created_at" binding:"omitempty,oneof=created_at full_name email role status"`
	SortDir  string `form:"sort_dir,default=desc" binding:"omitempty,oneof=asc desc"`
}

func (q AdminUserListRequest) URLParams() string {
	params := url.Values{}
	if q.Search != "" {
		params.Set("search", q.Search)
	}
	if q.Status != "" {
		params.Set("status", q.Status)
	}
	if q.Role != "" {
		params.Set("role", q.Role)
	}
	if q.SortBy != "" {
		params.Set("sort_by", q.SortBy)
	}
	if q.SortDir != "" {
		params.Set("sort_dir", q.SortDir)
	}
	return params.Encode()
}
