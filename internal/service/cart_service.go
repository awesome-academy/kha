package service

import (
	"errors"
	"fmt"

	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/models"
	"github.com/kha/foods-drinks/internal/repository"
	"gorm.io/gorm"
)

var (
	ErrCartNotFound      = errors.New("cart not found")
	ErrInsufficientStock = errors.New("insufficient stock")
	ErrInvalidQuantity   = errors.New("quantity must be at least 1")
)

type CartService struct {
	cartRepo    *repository.CartRepository
	productRepo *repository.ProductRepository
}

func NewCartService(cartRepo *repository.CartRepository, productRepo *repository.ProductRepository) *CartService {
	return &CartService{
		cartRepo:    cartRepo,
		productRepo: productRepo,
	}
}

func (s *CartService) EnsureCartForUser(userID uint) error {
	exists, err := s.cartRepo.ExistsByUserID(userID)
	if err != nil {
		return fmt.Errorf("failed to check cart: %w", err)
	}
	if exists {
		return nil
	}
	cart := &models.Cart{UserID: userID}
	return s.cartRepo.Create(cart)
}

func (s *CartService) GetCart(userID uint) (*dto.CartResponse, error) {
	cart, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cart = &models.Cart{UserID: userID}
			if err := s.cartRepo.Create(cart); err != nil {
				return nil, fmt.Errorf("failed to create cart: %w", err)
			}
			return s.toCartResponse(cart), nil
		}
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	return s.toCartResponse(cart), nil
}

func (s *CartService) AddItem(userID uint, req *dto.AddCartItemRequest) (*dto.CartResponse, error) {
	if req.Quantity < 1 {
		return nil, ErrInvalidQuantity
	}

	product, err := s.productRepo.FindByID(req.ProductID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	if product.Status != models.ProductStatusActive {
		return nil, ErrProductNotFound
	}

	if product.Stock < req.Quantity {
		return nil, fmt.Errorf("%w: available %d, requested %d", ErrInsufficientStock, product.Stock, req.Quantity)
	}

	cart, err := s.getOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	item, err := s.cartRepo.FindCartItem(cart.ID, req.ProductID)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("failed to find cart item: %w", err)
		}
		item = &models.CartItem{
			CartID:    cart.ID,
			ProductID: req.ProductID,
			Quantity:  req.Quantity,
		}
		if err := s.cartRepo.CreateCartItem(item); err != nil {
			return nil, fmt.Errorf("failed to add cart item: %w", err)
		}
	} else {
		newQty := item.Quantity + req.Quantity
		if product.Stock < newQty {
			return nil, fmt.Errorf("%w: available %d, requested total %d", ErrInsufficientStock, product.Stock, newQty)
		}
		item.Quantity = newQty
		if err := s.cartRepo.UpdateCartItem(item); err != nil {
			return nil, fmt.Errorf("failed to update cart item: %w", err)
		}
	}

	return s.GetCart(userID)
}

func (s *CartService) UpdateItem(userID uint, productID uint, quantity int) (*dto.CartResponse, error) {
	if quantity < 1 {
		return nil, ErrInvalidQuantity
	}

	product, err := s.productRepo.FindByID(productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}

	if product.Stock < quantity {
		return nil, fmt.Errorf("%w: available %d, requested %d", ErrInsufficientStock, product.Stock, quantity)
	}

	cart, err := s.getOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	item, err := s.cartRepo.FindCartItem(cart.ID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrCartNotFound
		}
		return nil, fmt.Errorf("failed to find cart item: %w", err)
	}

	item.Quantity = quantity
	if err := s.cartRepo.UpdateCartItem(item); err != nil {
		return nil, fmt.Errorf("failed to update cart item: %w", err)
	}

	return s.GetCart(userID)
}

func (s *CartService) RemoveItem(userID uint, productID uint) (*dto.CartResponse, error) {
	cart, err := s.getOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	if err := s.cartRepo.DeleteCartItem(cart.ID, productID); err != nil {
		return nil, fmt.Errorf("failed to remove cart item: %w", err)
	}

	return s.GetCart(userID)
}

func (s *CartService) ClearCart(userID uint) (*dto.CartResponse, error) {
	cart, err := s.getOrCreateCart(userID)
	if err != nil {
		return nil, err
	}

	if err := s.cartRepo.ClearCartItems(cart.ID); err != nil {
		return nil, fmt.Errorf("failed to clear cart: %w", err)
	}

	return s.GetCart(userID)
}

func (s *CartService) getOrCreateCart(userID uint) (*models.Cart, error) {
	cart, err := s.cartRepo.FindByUserID(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			cart = &models.Cart{UserID: userID}
			if err := s.cartRepo.Create(cart); err != nil {
				return nil, fmt.Errorf("failed to create cart: %w", err)
			}
			return cart, nil
		}
		return nil, fmt.Errorf("failed to get cart: %w", err)
	}
	return cart, nil
}

func (s *CartService) toCartResponse(cart *models.Cart) *dto.CartResponse {
	resp := &dto.CartResponse{
		ID:          cart.ID,
		Items:       make([]dto.CartItemResponse, 0, len(cart.Items)),
		TotalItems:  0,
		TotalAmount: 0,
	}

	for _, item := range cart.Items {
		subtotal := 0.0
		name := ""
		price := 0.0
		imageURL := ""

		if item.Product != nil {
			price = item.Product.Price
			name = item.Product.Name
			subtotal = price * float64(item.Quantity)
			if len(item.Product.Images) > 0 {
				for _, img := range item.Product.Images {
					if img.IsPrimary {
						imageURL = img.ImageURL
						break
					}
				}
				if imageURL == "" {
					imageURL = item.Product.Images[0].ImageURL
				}
			}
		}

		resp.Items = append(resp.Items, dto.CartItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Name:      name,
			Price:     price,
			Quantity:  item.Quantity,
			Subtotal:  subtotal,
			ImageURL:  imageURL,
		})
		resp.TotalItems += item.Quantity
		resp.TotalAmount += subtotal
	}

	return resp
}
