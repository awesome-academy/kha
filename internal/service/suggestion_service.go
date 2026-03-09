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
	ErrSuggestionNotFound     = errors.New("suggestion not found")
	ErrInvalidSuggestionState = errors.New("invalid suggestion status transition")
)

type SuggestionService struct {
	suggestionRepo *repository.SuggestionRepository
	categoryRepo   *repository.CategoryRepository
}

func NewSuggestionService(suggestionRepo *repository.SuggestionRepository, categoryRepo *repository.CategoryRepository) *SuggestionService {
	return &SuggestionService{suggestionRepo: suggestionRepo, categoryRepo: categoryRepo}
}

func (s *SuggestionService) Create(userID uint, req *dto.CreateSuggestionRequest) (*dto.SuggestionResponse, error) {
	suggestion := &models.Suggestion{
		UserID:   userID,
		Name:     strings.TrimSpace(req.Name),
		Classify: req.Classify,
		Status:   models.SuggestionStatusPending,
	}

	if req.Description != nil {
		d := strings.TrimSpace(*req.Description)
		if d != "" {
			suggestion.Description = &d
		}
	}

	if req.CategoryID != nil && *req.CategoryID > 0 {
		cat, err := s.categoryRepo.FindByID(*req.CategoryID)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil, ErrCategoryNotFound
			}
			return nil, fmt.Errorf("failed to find category: %w", err)
		}
		suggestion.CategoryID = &cat.ID
	}

	if err := s.suggestionRepo.Create(suggestion); err != nil {
		return nil, fmt.Errorf("failed to create suggestion: %w", err)
	}

	created, err := s.suggestionRepo.FindByID(suggestion.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to load suggestion: %w", err)
	}

	return toSuggestionResponse(created), nil
}

func (s *SuggestionService) ListForAdmin(req *dto.AdminSuggestionListRequest) (*dto.PaginatedResponse, error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 15
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}
	if req.SortDir == "" {
		req.SortDir = "desc"
	}

	offset := (req.Page - 1) * req.PageSize
	suggestions, total, err := s.suggestionRepo.ListForAdmin(repository.SuggestionListParams{
		Offset:   offset,
		Limit:    req.PageSize,
		Status:   req.Status,
		Classify: req.Classify,
		Search:   req.Search,
		SortBy:   req.SortBy,
		SortDir:  req.SortDir,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list suggestions: %w", err)
	}

	items := make([]dto.SuggestionResponse, len(suggestions))
	for i, suggestion := range suggestions {
		items[i] = *toSuggestionResponse(&suggestion)
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

func (s *SuggestionService) UpdateStatusForAdmin(id uint, req *dto.AdminUpdateSuggestionStatusRequest) error {
	status := strings.TrimSpace(req.Status)
	note := normalizeSuggestionNote(req.AdminNote)

	err := s.suggestionRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		txRepo := s.suggestionRepo.WithTx(tx)
		suggestion, err := txRepo.FindByIDForUpdate(id)
		if err != nil {
			if errors.Is(err, repository.ErrSuggestionNotFound) {
				return ErrSuggestionNotFound
			}
			return fmt.Errorf("failed to find suggestion: %w", err)
		}

		if suggestion.Status != models.SuggestionStatusPending {
			return ErrInvalidSuggestionState
		}
		if status != models.SuggestionStatusApproved && status != models.SuggestionStatusRejected {
			return ErrInvalidSuggestionState
		}

		suggestion.Status = status
		suggestion.AdminNote = note
		if err := txRepo.Update(suggestion); err != nil {
			return fmt.Errorf("failed to update suggestion: %w", err)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func toSuggestionResponse(suggestion *models.Suggestion) *dto.SuggestionResponse {
	resp := &dto.SuggestionResponse{
		ID:          suggestion.ID,
		UserID:      suggestion.UserID,
		Name:        suggestion.Name,
		Description: suggestion.Description,
		Classify:    suggestion.Classify,
		CategoryID:  suggestion.CategoryID,
		Status:      suggestion.Status,
		AdminNote:   suggestion.AdminNote,
		CreatedAt:   suggestion.CreatedAt,
		UpdatedAt:   suggestion.UpdatedAt,
	}
	if suggestion.User != nil {
		resp.UserName = suggestion.User.FullName
	}
	if suggestion.Category != nil {
		resp.CategoryName = suggestion.Category.Name
	}
	return resp
}

func normalizeSuggestionNote(note *string) *string {
	if note == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*note)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}
