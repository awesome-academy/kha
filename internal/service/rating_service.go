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
	ErrRatingNotFound      = errors.New("rating not found")
	ErrRatingAlreadyExists = errors.New("rating already exists")
	ErrProductNotPurchased = errors.New("product not purchased")
)

type RatingService struct {
	ratingRepo  *repository.RatingRepository
	productRepo *repository.ProductRepository
}

func NewRatingService(ratingRepo *repository.RatingRepository, productRepo *repository.ProductRepository) *RatingService {
	return &RatingService{ratingRepo: ratingRepo, productRepo: productRepo}
}

func (s *RatingService) CreateByProductSlug(userID uint, productSlug string, req *dto.CreateRatingRequest) (*dto.RatingResponse, error) {
	product, err := s.findProductBySlug(productSlug)
	if err != nil {
		return nil, err
	}
	productID := product.ID

	orderID, err := s.ratingRepo.FindPurchasedOrderID(userID, productID)
	if err != nil {
		return nil, fmt.Errorf("failed to verify purchase: %w", err)
	}
	if orderID == nil {
		return nil, ErrProductNotPurchased
	}

	existing, err := s.ratingRepo.FindByUserAndProduct(userID, productID)
	if err == nil && existing != nil {
		return nil, ErrRatingAlreadyExists
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, fmt.Errorf("failed to check existing rating: %w", err)
	}

	comment := normalizeComment(req.Comment)
	rating := &models.Rating{
		UserID:    userID,
		ProductID: productID,
		OrderID:   orderID,
		Rating:    req.Rating,
		Comment:   comment,
	}

	err = s.ratingRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		ratingRepoTx := s.ratingRepo.WithTx(tx)

		if err := ratingRepoTx.Create(rating); err != nil {
			if isDuplicateRatingError(err) {
				return ErrRatingAlreadyExists
			}
			return fmt.Errorf("failed to create rating: %w", err)
		}

		avg, count, err := ratingRepoTx.CalcProductRatingStats(productID)
		if err != nil {
			return fmt.Errorf("failed to calculate rating stats: %w", err)
		}
		if err := ratingRepoTx.UpdateProductRatingStats(productID, avg, count); err != nil {
			return fmt.Errorf("failed to update product rating stats: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.RatingResponse{
		ID:        rating.ID,
		UserID:    userID,
		ProductID: productID,
		OrderID:   orderID,
		Rating:    rating.Rating,
		Comment:   rating.Comment,
		CreatedAt: rating.CreatedAt,
		UpdatedAt: rating.UpdatedAt,
	}, nil
}

func (s *RatingService) UpdateByProductSlug(userID uint, productSlug string, req *dto.UpdateRatingRequest) (*dto.RatingResponse, error) {
	product, err := s.findProductBySlug(productSlug)
	if err != nil {
		return nil, err
	}
	productID := product.ID

	rating, err := s.ratingRepo.FindByUserAndProduct(userID, productID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRatingNotFound
		}
		return nil, fmt.Errorf("failed to find rating: %w", err)
	}

	rating.Rating = req.Rating
	rating.Comment = normalizeComment(req.Comment)

	err = s.ratingRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		ratingRepoTx := s.ratingRepo.WithTx(tx)

		if err := ratingRepoTx.Update(rating); err != nil {
			return fmt.Errorf("failed to update rating: %w", err)
		}

		avg, count, err := ratingRepoTx.CalcProductRatingStats(rating.ProductID)
		if err != nil {
			return fmt.Errorf("failed to calculate rating stats: %w", err)
		}
		if err := ratingRepoTx.UpdateProductRatingStats(rating.ProductID, avg, count); err != nil {
			return fmt.Errorf("failed to update product rating stats: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return &dto.RatingResponse{
		ID:        rating.ID,
		UserID:    rating.UserID,
		ProductID: rating.ProductID,
		OrderID:   rating.OrderID,
		Rating:    rating.Rating,
		Comment:   rating.Comment,
		CreatedAt: rating.CreatedAt,
		UpdatedAt: rating.UpdatedAt,
	}, nil
}

func (s *RatingService) ListByProductSlug(productSlug string, req *dto.RatingListRequest) (*dto.PaginatedResponse, error) {
	product, err := s.findProductBySlug(productSlug)
	if err != nil {
		return nil, err
	}
	productID := product.ID

	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 20
	}

	offset := (req.Page - 1) * req.PageSize
	ratings, total, err := s.ratingRepo.ListByProductID(productID, offset, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list ratings: %w", err)
	}

	items := make([]dto.RatingResponse, len(ratings))
	for i, rating := range ratings {
		items[i] = dto.RatingResponse{
			ID:        rating.ID,
			UserID:    rating.UserID,
			ProductID: rating.ProductID,
			OrderID:   rating.OrderID,
			Rating:    rating.Rating,
			Comment:   rating.Comment,
			CreatedAt: rating.CreatedAt,
			UpdatedAt: rating.UpdatedAt,
		}
		if rating.User != nil {
			items[i].UserName = rating.User.FullName
			items[i].UserAvatar = rating.User.AvatarURL
		}
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

func normalizeComment(comment *string) *string {
	if comment == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*comment)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

func isDuplicateRatingError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate") || strings.Contains(message, "1062")
}

func (s *RatingService) findProductBySlug(productSlug string) (*models.Product, error) {
	product, err := s.productRepo.FindBySlug(strings.TrimSpace(productSlug))
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrProductNotFound
		}
		return nil, fmt.Errorf("failed to find product: %w", err)
	}
	if product.Status != models.ProductStatusActive {
		return nil, ErrProductNotFound
	}
	return product, nil
}
