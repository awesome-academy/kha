package repository

import (
	"errors"
	"strings"

	"github.com/kha/foods-drinks/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var ErrSuggestionNotFound = errors.New("suggestion not found")

type SuggestionRepository struct {
	db *gorm.DB
}

func NewSuggestionRepository(db *gorm.DB) *SuggestionRepository {
	return &SuggestionRepository{db: db}
}

func (r *SuggestionRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *SuggestionRepository) WithTx(tx *gorm.DB) *SuggestionRepository {
	return &SuggestionRepository{db: tx}
}

func (r *SuggestionRepository) Create(suggestion *models.Suggestion) error {
	return r.db.Create(suggestion).Error
}

func (r *SuggestionRepository) FindByID(id uint) (*models.Suggestion, error) {
	var suggestion models.Suggestion
	err := r.db.Where("id = ?", id).
		Preload("User").
		Preload("Category").
		First(&suggestion).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSuggestionNotFound
		}
		return nil, err
	}
	return &suggestion, nil
}

func (r *SuggestionRepository) FindByIDForUpdate(id uint) (*models.Suggestion, error) {
	var suggestion models.Suggestion
	err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", id).First(&suggestion).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrSuggestionNotFound
		}
		return nil, err
	}
	return &suggestion, nil
}

type SuggestionListParams struct {
	Offset   int
	Limit    int
	Status   string
	Classify string
	Search   string
	SortBy   string
	SortDir  string
}

func (r *SuggestionRepository) ListForAdmin(params SuggestionListParams) ([]models.Suggestion, int64, error) {
	var suggestions []models.Suggestion
	var total int64

	query := r.db.Model(&models.Suggestion{})
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Classify != "" {
		query = query.Where("classify = ?", params.Classify)
	}
	if params.Search != "" {
		like := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("name LIKE ?", like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "created_at"
	if params.SortBy == "status" || params.SortBy == "name" {
		sortBy = params.SortBy
	}
	sortDir := "desc"
	if params.SortDir == "asc" {
		sortDir = "asc"
	}

	err := query.
		Preload("User").
		Preload("Category").
		Order(sortBy + " " + sortDir).
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&suggestions).Error
	if err != nil {
		return nil, 0, err
	}

	return suggestions, total, nil
}

func (r *SuggestionRepository) Update(suggestion *models.Suggestion) error {
	return r.db.Save(suggestion).Error
}
