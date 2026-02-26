package service

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
	"github.com/kha/foods-drinks/internal/config"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrNoAvatar        = errors.New("no avatar to delete")
	ErrFileTooLarge    = errors.New("file too large")
	ErrInvalidFileType = errors.New("invalid file type")
)

// ProfileService handles user profile operations (avatar upload/delete)
type ProfileService struct {
	userRepo     *repository.UserRepository
	uploadConfig *config.UploadConfig
}

// NewProfileService creates a new ProfileService
func NewProfileService(userRepo *repository.UserRepository, uploadConfig *config.UploadConfig) *ProfileService {
	return &ProfileService{
		userRepo:     userRepo,
		uploadConfig: uploadConfig,
	}
}

// UploadAvatar uploads an avatar image for a user
func (s *ProfileService) UploadAvatar(userID uint, file *multipart.FileHeader) (*dto.AvatarResponse, error) {
	// Validate file size
	if file.Size > s.uploadConfig.MaxSize {
		return nil, ErrFileTooLarge
	}

	// Validate file type
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !s.isAllowedType(ext) {
		return nil, ErrInvalidFileType
	}

	// Find user
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Delete old avatar file if it exists
	if user.AvatarURL != nil && *user.AvatarURL != "" {
		oldPath := strings.TrimPrefix(*user.AvatarURL, "/")
		_ = os.Remove(oldPath) // best-effort removal
	}

	// Generate unique filename
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	savePath := filepath.Join(s.uploadConfig.Path, filename)

	// Ensure upload directory exists
	if err := os.MkdirAll(s.uploadConfig.Path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Open uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer func() { _ = src.Close() }()

	// Create destination file
	dst, err := os.Create(savePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer func() { _ = dst.Close() }()

	// Copy file content
	if _, err := io.Copy(dst, src); err != nil {
		_ = os.Remove(savePath) // clean up on failure
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	// Update user avatar URL
	avatarURL := "/" + savePath
	user.AvatarURL = &avatarURL

	if err := s.userRepo.Update(user); err != nil {
		_ = os.Remove(savePath) // clean up on failure
		return nil, fmt.Errorf("failed to update user avatar: %w", err)
	}

	return &dto.AvatarResponse{
		AvatarURL: avatarURL,
	}, nil
}

// DeleteAvatar removes the avatar for a user
func (s *ProfileService) DeleteAvatar(userID uint) error {
	user, err := s.userRepo.FindByID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrUserNotFound
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user.AvatarURL == nil || *user.AvatarURL == "" {
		return ErrNoAvatar
	}

	// Delete the file from disk
	oldPath := strings.TrimPrefix(*user.AvatarURL, "/")
	_ = os.Remove(oldPath) // best-effort removal

	// Clear avatar URL in DB
	user.AvatarURL = nil
	if err := s.userRepo.Update(user); err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	return nil
}

// isAllowedType checks if the file extension is in the allowed types list
func (s *ProfileService) isAllowedType(ext string) bool {
	for _, allowed := range s.uploadConfig.AllowedTypes {
		if ext == "."+allowed {
			return true
		}
	}
	return false
}
