package repository

import (
	"time"

	"github.com/kha/foods-drinks/internal/models"
	"gorm.io/gorm"
)

type OrderRepository struct {
	db *gorm.DB
}

func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) WithTx(tx *gorm.DB) *OrderRepository {
	return &OrderRepository{db: tx}
}

func (r *OrderRepository) Create(order *models.Order) error {
	return r.db.Create(order).Error
}

func (r *OrderRepository) CreateItems(items []models.OrderItem) error {
	if len(items) == 0 {
		return nil
	}
	return r.db.Create(&items).Error
}

func (r *OrderRepository) FindByIDAndUserID(orderID uint, userID uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Where("id = ? AND user_id = ?", orderID, userID).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_items.id ASC")
		}).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

type OrderListParams struct {
	Offset   int
	Limit    int
	UserID   uint
	Status   string
	FromDate *time.Time
	ToDate   *time.Time
}

type AdminOrderListParams struct {
	Offset   int
	Limit    int
	Status   string
	FromDate *time.Time
	ToDate   *time.Time
	SortBy   string
	SortDir  string
}

func (r *OrderRepository) ListByUserID(params OrderListParams) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{}).Where("user_id = ?", params.UserID)
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.FromDate != nil {
		query = query.Where("created_at >= ?", *params.FromDate)
	}
	if params.ToDate != nil {
		query = query.Where("created_at <= ?", *params.ToDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_items.id ASC")
		}).
		Order("created_at DESC").
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *OrderRepository) ListForAdmin(params AdminOrderListParams) ([]models.Order, int64, error) {
	var orders []models.Order
	var total int64

	query := r.db.Model(&models.Order{})
	if params.Status != "" {
		query = query.Where("status = ?", params.Status)
	}
	if params.FromDate != nil {
		query = query.Where("created_at >= ?", *params.FromDate)
	}
	if params.ToDate != nil {
		query = query.Where("created_at <= ?", *params.ToDate)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	sortBy := "created_at"
	if params.SortBy == "total_amount" || params.SortBy == "status" {
		sortBy = params.SortBy
	}
	sortDir := "desc"
	if params.SortDir == "asc" {
		sortDir = "asc"
	}

	err := query.
		Preload("User").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_items.id ASC")
		}).
		Order(sortBy + " " + sortDir).
		Offset(params.Offset).
		Limit(params.Limit).
		Find(&orders).Error
	if err != nil {
		return nil, 0, err
	}

	return orders, total, nil
}

func (r *OrderRepository) FindByID(id uint) (*models.Order, error) {
	var order models.Order
	err := r.db.Where("id = ?", id).
		Preload("User").
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("order_items.id ASC")
		}).
		First(&order).Error
	if err != nil {
		return nil, err
	}
	return &order, nil
}

func (r *OrderRepository) Update(order *models.Order) error {
	return r.db.Save(order).Error
}
