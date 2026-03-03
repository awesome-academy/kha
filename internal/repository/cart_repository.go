package repository

import (
	"github.com/kha/foods-drinks/internal/models"
	"gorm.io/gorm"
)

type CartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) *CartRepository {
	return &CartRepository{db: db}
}

func (r *CartRepository) GetDB() *gorm.DB {
	return r.db
}

func (r *CartRepository) WithTx(tx *gorm.DB) *CartRepository {
	return &CartRepository{db: tx}
}

func (r *CartRepository) Create(cart *models.Cart) error {
	return r.db.Create(cart).Error
}

func (r *CartRepository) FindByUserID(userID uint) (*models.Cart, error) {
	var cart models.Cart
	err := r.db.Where("user_id = ?", userID).
		Preload("Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("cart_items.created_at ASC")
		}).
		Preload("Items.Product").Preload("Items.Product.Images").
		First(&cart).Error
	if err != nil {
		return nil, err
	}
	return &cart, nil
}

func (r *CartRepository) ExistsByUserID(userID uint) (bool, error) {
	var count int64
	err := r.db.Model(&models.Cart{}).Where("user_id = ?", userID).Count(&count).Error
	return count > 0, err
}

func (r *CartRepository) FindCartItem(cartID, productID uint) (*models.CartItem, error) {
	var item models.CartItem
	err := r.db.Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error
	if err != nil {
		return nil, err
	}
	return &item, nil
}

func (r *CartRepository) CreateCartItem(item *models.CartItem) error {
	return r.db.Create(item).Error
}

func (r *CartRepository) UpdateCartItem(item *models.CartItem) error {
	return r.db.Save(item).Error
}

func (r *CartRepository) DeleteCartItem(cartID, productID uint) error {
	return r.db.Where("cart_id = ? AND product_id = ?", cartID, productID).
		Delete(&models.CartItem{}).Error
}

func (r *CartRepository) ClearCartItems(cartID uint) error {
	return r.db.Where("cart_id = ?", cartID).Delete(&models.CartItem{}).Error
}
