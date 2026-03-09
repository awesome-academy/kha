package repository

import (
	"github.com/kha/foods-drinks/internal/models"
	"gorm.io/gorm"
)

type RatingRepository struct {
	db *gorm.DB
}

func NewRatingRepository(db *gorm.DB) *RatingRepository {
	return &RatingRepository{db: db}
}

func (r *RatingRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *RatingRepository) WithTx(tx *gorm.DB) *RatingRepository {
	return &RatingRepository{db: tx}
}

func (r *RatingRepository) Create(rating *models.Rating) error {
	return r.db.Create(rating).Error
}

func (r *RatingRepository) Update(rating *models.Rating) error {
	return r.db.Save(rating).Error
}

func (r *RatingRepository) FindByUserAndProduct(userID, productID uint) (*models.Rating, error) {
	var rating models.Rating
	err := r.db.Where("user_id = ? AND product_id = ?", userID, productID).First(&rating).Error
	if err != nil {
		return nil, err
	}
	return &rating, nil
}

func (r *RatingRepository) ListByProductID(productID uint, offset, limit int) ([]models.Rating, int64, error) {
	var ratings []models.Rating
	var total int64

	query := r.db.Model(&models.Rating{}).Where("product_id = ?", productID)
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	err := query.
		Preload("User").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&ratings).Error
	if err != nil {
		return nil, 0, err
	}

	return ratings, total, nil
}

func (r *RatingRepository) FindPurchasedOrderID(userID, productID uint) (*uint, error) {
	var orderID uint
	err := r.db.Table("orders").
		Select("orders.id").
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Where("orders.user_id = ? AND order_items.product_id = ? AND orders.status <> ?", userID, productID, models.OrderStatusCancelled).
		Order("orders.created_at DESC").
		Limit(1).
		Scan(&orderID).Error
	if err != nil {
		return nil, err
	}
	if orderID == 0 {
		return nil, nil
	}
	return &orderID, nil
}

func (r *RatingRepository) CalcProductRatingStats(productID uint) (float64, int64, error) {
	type stats struct {
		Average float64
		Count   int64
	}

	var s stats
	err := r.db.Model(&models.Rating{}).
		Select("COALESCE(AVG(rating), 0) as average, COUNT(*) as count").
		Where("product_id = ?", productID).
		Scan(&s).Error
	if err != nil {
		return 0, 0, err
	}

	return s.Average, s.Count, nil
}

func (r *RatingRepository) UpdateProductRatingStats(productID uint, average float64, count int64) error {
	return r.db.Model(&models.Product{}).
		Where("id = ?", productID).
		Updates(map[string]interface{}{
			"rating_average": average,
			"rating_count":   count,
		}).Error
}
