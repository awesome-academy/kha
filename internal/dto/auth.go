package dto

import "time"

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email,max=255"`
	Password string `json:"password" binding:"required,min=8,max=72"`
	FullName string `json:"full_name" binding:"required,min=2,max=255"`
}

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse represents the response for successful authentication
type AuthResponse struct {
	AccessToken string       `json:"access_token"`
	TokenType   string       `json:"token_type"`
	ExpiresIn   int64        `json:"expires_in"`
	User        UserResponse `json:"user"`
}

// UserResponse represents user data in response
type UserResponse struct {
	ID              uint       `json:"id"`
	Email           string     `json:"email"`
	FullName        string     `json:"full_name"`
	Phone           *string    `json:"phone,omitempty"`
	Address         *string    `json:"address,omitempty"`
	AvatarURL       *string    `json:"avatar_url,omitempty"`
	Role            string     `json:"role"`
	Status          string     `json:"status"`
	EmailVerifiedAt *time.Time `json:"email_verified_at,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string            `json:"error"`
	Message string            `json:"message"`
	Details map[string]string `json:"details,omitempty"`
}

// UpdateProfileRequest represents the request body for updating user profile
type UpdateProfileRequest struct {
	FullName string  `json:"full_name" binding:"required,min=2,max=255"`
	Phone    *string `json:"phone" binding:"omitempty,min=10,max=20"`
	Address  *string `json:"address" binding:"omitempty,max=500"`
}

// OAuthURLResponse represents the response for OAuth URL request
type OAuthURLResponse struct {
	URL      string `json:"url"`
	Provider string `json:"provider"`
	State    string `json:"state"`
}

// OAuthCallbackRequest represents the OAuth callback query parameters
type OAuthCallbackRequest struct {
	Code  string `form:"code" binding:"required"`
	State string `form:"state" binding:"required"`
}

// OAuthProvidersResponse represents the list of supported OAuth providers
type OAuthProvidersResponse struct {
	Providers []string `json:"providers"`
}
