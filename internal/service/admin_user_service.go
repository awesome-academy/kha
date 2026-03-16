package service

import (
	"errors"
	"fmt"
	"math"
	"strings"

	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/models"
	"github.com/kha/foods-drinks/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrAdminUserNotFound = errors.New("admin user not found")
	ErrInvalidUserRole   = errors.New("invalid user role")
	ErrInvalidUserStatus = errors.New("invalid user status")
	ErrCannotBanAdmin    = errors.New("cannot ban admin user")
	ErrCannotBanSelf     = errors.New("cannot ban yourself")
)

type AdminUserService struct {
	userRepo *repository.UserRepository
}

func NewAdminUserService(userRepo *repository.UserRepository) *AdminUserService {
	return &AdminUserService{userRepo: userRepo}
}

func (s *AdminUserService) ListForAdmin(req *dto.AdminUserListRequest) (*dto.PaginatedResponse, error) {
	if req == nil {
		req = &dto.AdminUserListRequest{}
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 15
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	offset := (req.Page - 1) * req.PageSize
	users, total, err := s.userRepo.ListForAdmin(repository.AdminUserListParams{
		Offset:  offset,
		Limit:   req.PageSize,
		Search:  strings.TrimSpace(req.Search),
		Status:  strings.TrimSpace(req.Status),
		Role:    strings.TrimSpace(req.Role),
		SortBy:  strings.TrimSpace(req.SortBy),
		SortDir: strings.TrimSpace(req.SortDir),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	items := make([]dto.UserResponse, 0, len(users))
	for i := range users {
		items = append(items, dto.ToUserResponse(&users[i]))
	}

	totalPages := int(math.Ceil(float64(total) / float64(req.PageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	return &dto.PaginatedResponse{
		Items:      items,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}, nil
}

func (s *AdminUserService) GetDetailForAdmin(id uint) (*dto.UserResponse, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrAdminUserNotFound
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	resp := dto.ToUserResponse(user)
	return &resp, nil
}

func (s *AdminUserService) UpdateStatusForAdmin(targetUserID uint, status string, actorUserID uint) error {
	status = strings.TrimSpace(status)
	if !isValidUserStatus(status) {
		return ErrInvalidUserStatus
	}

	return s.userRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.userRepo.WithTx(tx)
		user, err := txRepo.FindByIDForUpdate(targetUserID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrAdminUserNotFound
			}
			return fmt.Errorf("failed to find user: %w", err)
		}

		if user.Role == models.RoleAdmin && status == models.UserStatusBanned {
			return ErrCannotBanAdmin
		}
		if actorUserID != 0 && actorUserID == user.ID && status == models.UserStatusBanned {
			return ErrCannotBanSelf
		}

		if user.Status == status {
			return nil
		}

		if err := txRepo.UpdateFields(user.ID, map[string]interface{}{"status": status}); err != nil {
			return fmt.Errorf("failed to update user status: %w", err)
		}

		return nil
	})
}

func (s *AdminUserService) UpdateRoleForAdmin(targetUserID uint, role string) error {
	role = strings.TrimSpace(role)
	if !isValidUserRole(role) {
		return ErrInvalidUserRole
	}

	user, err := s.userRepo.FindByID(targetUserID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrAdminUserNotFound
		}
		return fmt.Errorf("failed to find user: %w", err)
	}

	if user.Role == role {
		return nil
	}

	if err := s.userRepo.UpdateFields(user.ID, map[string]interface{}{"role": role}); err != nil {
		return fmt.Errorf("failed to update user role: %w", err)
	}

	return nil
}

func isValidUserRole(role string) bool {
	return role == models.RoleUser || role == models.RoleAdmin
}

func isValidUserStatus(status string) bool {
	return status == models.UserStatusActive || status == models.UserStatusInactive || status == models.UserStatusBanned
}
