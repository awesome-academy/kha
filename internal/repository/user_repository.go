package repository

import (
	"strings"

	"github.com/kha/foods-drinks/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *gorm.DB
}

type AdminUserListParams struct {
	Offset  int
	Limit   int
	Search  string
	Status  string
	Role    string
	SortBy  string
	SortDir string
}

// NewUserRepository creates a new UserRepository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

// GetDB returns the underlying database connection for transaction support
func (r *UserRepository) GetDB() *gorm.DB {
	return r.db
}

// WithTx returns a new UserRepository that uses the given transaction
func (r *UserRepository) WithTx(tx *gorm.DB) *UserRepository {
	return &UserRepository{db: tx}
}

// Create creates a new user
func (r *UserRepository) Create(user *models.User) error {
	return r.db.Create(user).Error
}

// FindByID finds a user by ID
func (r *UserRepository) FindByID(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByIDForUpdate finds user row with FOR UPDATE lock.
func (r *UserRepository) FindByIDForUpdate(id uint) (*models.User, error) {
	var user models.User
	if err := r.db.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User
	if err := r.db.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// ExistsByEmail checks if a user with the given email exists
func (r *UserRepository) ExistsByEmail(email string) (bool, error) {
	var count int64
	if err := r.db.Model(&models.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

// Update updates a user
func (r *UserRepository) Update(user *models.User) error {
	return r.db.Save(user).Error
}

// UpdateFields updates specific fields of a user
func (r *UserRepository) UpdateFields(id uint, fields map[string]interface{}) error {
	return r.db.Model(&models.User{}).Where("id = ?", id).Updates(fields).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&models.User{}, id).Error
}

// List returns a list of users with pagination
func (r *UserRepository) List(offset, limit int) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	if err := r.db.Model(&models.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	if err := r.db.Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// ListForAdmin returns users with filtering/sorting/pagination for admin SSR.
func (r *UserRepository) ListForAdmin(params AdminUserListParams) ([]models.User, int64, error) {
	var users []models.User
	var total int64

	query := r.db.Model(&models.User{})

	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.Role != "" {
		query = query.Where("role = ?", params.Role)
	}
	if params.Search != "" {
		like := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("full_name LIKE ? OR email LIKE ?", like, like)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "created_at"
	switch params.SortBy {
	case "full_name", "email", "role", "status", "created_at":
		sortBy = params.SortBy
	}
	sortDir := "desc"
	if strings.EqualFold(params.SortDir, "asc") {
		sortDir = "asc"
	}

	if err := query.Order(sortBy + " " + sortDir).Offset(params.Offset).Limit(params.Limit).Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// FindByIDWithRelations finds a user by ID with specified relations
// Only allows whitelisted relations to prevent performance issues and unexpected behavior
func (r *UserRepository) FindByIDWithRelations(id uint, relations ...string) (*models.User, error) {
	// Whitelist of allowed relations for User model
	allowedRelations := map[string]bool{
		"SocialAuths": true,
		"Cart":        true,
		"Orders":      true,
		"Ratings":     true,
	}

	var user models.User
	query := r.db
	for _, rel := range relations {
		// Only preload if relation is in whitelist
		if allowedRelations[rel] {
			query = query.Preload(rel)
		}
	}
	if err := query.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}
