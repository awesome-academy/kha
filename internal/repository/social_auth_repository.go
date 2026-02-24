package repository

import (
	"github.com/kha/foods-drinks/internal/models"
	"gorm.io/gorm"
)

// SocialAuthRepository handles social auth database operations
type SocialAuthRepository struct {
	db *gorm.DB
}

// NewSocialAuthRepository creates a new SocialAuthRepository
func NewSocialAuthRepository(db *gorm.DB) *SocialAuthRepository {
	return &SocialAuthRepository{db: db}
}

// Create creates a new social auth record
func (r *SocialAuthRepository) Create(socialAuth *models.SocialAuth) error {
	return r.db.Create(socialAuth).Error
}

// FindByProviderAndUserID finds a social auth by provider and provider user ID
func (r *SocialAuthRepository) FindByProviderAndProviderUserID(provider, providerUserID string) (*models.SocialAuth, error) {
	var socialAuth models.SocialAuth
	if err := r.db.Where("provider = ? AND provider_user_id = ?", provider, providerUserID).First(&socialAuth).Error; err != nil {
		return nil, err
	}
	return &socialAuth, nil
}

// FindByUserIDAndProvider finds a social auth by user ID and provider
func (r *SocialAuthRepository) FindByUserIDAndProvider(userID uint, provider string) (*models.SocialAuth, error) {
	var socialAuth models.SocialAuth
	if err := r.db.Where("user_id = ? AND provider = ?", userID, provider).First(&socialAuth).Error; err != nil {
		return nil, err
	}
	return &socialAuth, nil
}

// FindByUserID finds all social auths for a user
func (r *SocialAuthRepository) FindByUserID(userID uint) ([]models.SocialAuth, error) {
	var socialAuths []models.SocialAuth
	if err := r.db.Where("user_id = ?", userID).Find(&socialAuths).Error; err != nil {
		return nil, err
	}
	return socialAuths, nil
}

// Update updates a social auth record
func (r *SocialAuthRepository) Update(socialAuth *models.SocialAuth) error {
	return r.db.Save(socialAuth).Error
}

// UpdateTokens updates the access and refresh tokens
func (r *SocialAuthRepository) UpdateTokens(id uint, accessToken, refreshToken *string) error {
	return r.db.Model(&models.SocialAuth{}).Where("id = ?", id).Updates(map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	}).Error
}

// Delete deletes a social auth record
func (r *SocialAuthRepository) Delete(id uint) error {
	return r.db.Delete(&models.SocialAuth{}, id).Error
}

// DeleteByUserIDAndProvider deletes a social auth by user ID and provider
func (r *SocialAuthRepository) DeleteByUserIDAndProvider(userID uint, provider string) error {
	return r.db.Where("user_id = ? AND provider = ?", userID, provider).Delete(&models.SocialAuth{}).Error
}
