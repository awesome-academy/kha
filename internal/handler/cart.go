package handler

import (
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/middleware"
	"github.com/kha/foods-drinks/internal/service"
)

type CartHandler struct {
	cartService *service.CartService
}

func NewCartHandler(cartService *service.CartService) *CartHandler {
	return &CartHandler{cartService: cartService}
}

// Get godoc
// @Summary Get current user cart
// @Description Get cart of currently authenticated user
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.CartResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/cart [get]
func (h *CartHandler) Get(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	cart, err := h.cartService.GetCart(userID)
	if err != nil {
		log.Printf("Cart get error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to get cart",
		})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// Add godoc
// @Summary Add item to cart
// @Description Add product to current user cart
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AddCartItemRequest true "Add cart item request"
// @Success 200 {object} dto.CartResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/cart/items [post]
func (h *CartHandler) Add(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req dto.AddCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request: " + err.Error(),
		})
		return
	}

	cart, err := h.cartService.AddItem(userID, &req)
	if err != nil {
		h.handleCartError(c, err)
		return
	}

	c.JSON(http.StatusOK, cart)
}

// Update godoc
// @Summary Update item quantity in cart
// @Description Update quantity for a product in current user cart
// @Tags cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param product_id path int true "Product ID"
// @Param request body dto.UpdateCartItemRequest true "Update cart item request"
// @Success 200 {object} dto.CartResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/cart/items/{product_id} [put]
func (h *CartHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	productID, ok := parsePositiveUintParam(c.Param("product_id"))
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_product_id",
			Message: "Invalid product ID",
		})
		return
	}

	var req dto.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Invalid request: quantity is required and must be at least 1",
		})
		return
	}

	cart, err := h.cartService.UpdateItem(userID, uint(productID), req.Quantity)
	if err != nil {
		h.handleCartError(c, err)
		return
	}

	c.JSON(http.StatusOK, cart)
}

// Remove godoc
// @Summary Remove item from cart
// @Description Remove a product from current user cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Param product_id path int true "Product ID"
// @Success 200 {object} dto.CartResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/cart/items/{product_id} [delete]
func (h *CartHandler) Remove(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	productID, ok := parsePositiveUintParam(c.Param("product_id"))
	if !ok {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_product_id",
			Message: "Invalid product ID",
		})
		return
	}

	cart, err := h.cartService.RemoveItem(userID, uint(productID))
	if err != nil {
		h.handleCartError(c, err)
		return
	}

	c.JSON(http.StatusOK, cart)
}

// Clear godoc
// @Summary Clear current user cart
// @Description Remove all items in current user cart
// @Tags cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.CartResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/cart [delete]
func (h *CartHandler) Clear(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	cart, err := h.cartService.ClearCart(userID)
	if err != nil {
		log.Printf("Cart clear error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "Failed to clear cart",
		})
		return
	}

	c.JSON(http.StatusOK, cart)
}

func (h *CartHandler) handleCartError(c *gin.Context, err error) {
	respond := func(status int, code, fallbackMessage string) {
		c.JSON(status, dto.ErrorResponse{
			Error:   code,
			Message: fallbackMessage,
		})
	}

	switch {
	case errors.Is(err, service.ErrProductNotFound):
		respond(http.StatusNotFound, "product_not_found", "Product not found")
	case errors.Is(err, service.ErrInsufficientStock):
		message := "Insufficient stock"
		if err != nil && err.Error() != "" {
			message = err.Error()
		}
		respond(http.StatusBadRequest, "insufficient_stock", message)
	case errors.Is(err, service.ErrInvalidQuantity):
		respond(http.StatusBadRequest, "invalid_quantity", "Quantity must be at least 1")
	case errors.Is(err, service.ErrCartItemNotFound):
		respond(http.StatusNotFound, "cart_item_not_found", "Item not found in cart")
	case errors.Is(err, service.ErrCartNotFound):
		respond(http.StatusNotFound, "cart_not_found", "Cart not found")
	default:
		log.Printf("Cart error: %v", err)
		respond(http.StatusInternalServerError, "internal_error", "An unexpected error occurred")
	}
}

func parsePositiveUintParam(raw string) (uint64, bool) {
	value := strings.TrimSpace(raw)
	if value == "" || len(value) > 20 {
		return 0, false
	}

	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return 0, false
		}
	}

	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil || id == 0 {
		return 0, false
	}

	return id, true
}
