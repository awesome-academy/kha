package dto

import (
	"github.com/kha/foods-drinks/internal/models"
)

// ToUserResponse converts a User model to UserResponse DTO
// This is a shared utility function to avoid code duplication across services
func ToUserResponse(user *models.User) UserResponse {
	return UserResponse{
		ID:              user.ID,
		Email:           user.Email,
		FullName:        user.FullName,
		Phone:           user.Phone,
		Address:         user.Address,
		AvatarURL:       user.AvatarURL,
		Role:            user.Role,
		Status:          user.Status,
		EmailVerifiedAt: user.EmailVerifiedAt,
		CreatedAt:       user.CreatedAt,
		UpdatedAt:       user.UpdatedAt,
	}
}

// AvatarResponse returned after avatar upload
type AvatarResponse struct {
	AvatarURL string `json:"avatar_url"`
}
