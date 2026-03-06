package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/middleware"
	"github.com/kha/foods-drinks/internal/service"
)

type RatingHandler struct {
	ratingService *service.RatingService
}

func NewRatingHandler(ratingService *service.RatingService) *RatingHandler {
	return &RatingHandler{ratingService: ratingService}
}

// ListByProduct godoc
// @Summary List product ratings
// @Description Public API list ratings of a product with user info
// @Tags ratings
// @Produce json
// @Param slug path string true "Product slug"
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {object} dto.PaginatedResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/products/{slug}/ratings [get]
func (h *RatingHandler) ListByProduct(c *gin.Context) {
	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_slug", Message: "Invalid product slug"})
		return
	}

	var req dto.RatingListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_params", Message: "Invalid query parameters: " + err.Error()})
		return
	}

	resp, err := h.ratingService.ListByProductSlug(slug, &req)
	if err != nil {
		h.handleRatingError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// Create godoc
// @Summary Create product rating
// @Description Create a rating for product by authenticated user (must have purchased)
// @Tags ratings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Product slug"
// @Param request body dto.CreateRatingRequest true "Create rating request"
// @Success 201 {object} dto.RatingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/products/{slug}/ratings [post]
func (h *RatingHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_slug", Message: "Invalid product slug"})
		return
	}

	var req dto.CreateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: "Invalid request: " + err.Error()})
		return
	}

	resp, err := h.ratingService.CreateByProductSlug(userID, slug, &req)
	if err != nil {
		h.handleRatingError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Update godoc
// @Summary Update product rating
// @Description Update rating of authenticated user for a product
// @Tags ratings
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param slug path string true "Product slug"
// @Param request body dto.UpdateRatingRequest true "Update rating request"
// @Success 200 {object} dto.RatingResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/products/{slug}/ratings [put]
func (h *RatingHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	slug := strings.TrimSpace(c.Param("slug"))
	if slug == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_slug", Message: "Invalid product slug"})
		return
	}

	var req dto.UpdateRatingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: "Invalid request: " + err.Error()})
		return
	}

	resp, err := h.ratingService.UpdateByProductSlug(userID, slug, &req)
	if err != nil {
		h.handleRatingError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *RatingHandler) handleRatingError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrProductNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "product_not_found", Message: "Product not found"})
	case errors.Is(err, service.ErrProductNotPurchased):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "product_not_purchased", Message: "You can only rate products you have purchased"})
	case errors.Is(err, service.ErrRatingAlreadyExists):
		c.JSON(http.StatusConflict, dto.ErrorResponse{Error: "rating_exists", Message: "You have already rated this product"})
	case errors.Is(err, service.ErrRatingNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{Error: "rating_not_found", Message: "Rating not found"})
	default:
		log.Printf("Rating error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "An unexpected error occurred"})
	}
}
