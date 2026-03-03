package handler

import (
	"log"
	"net/http"
	"strconv"

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

func (h *CartHandler) Update(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 32)
	if err != nil || productID == 0 {
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

func (h *CartHandler) Remove(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	productID, err := strconv.ParseUint(c.Param("product_id"), 10, 32)
	if err != nil || productID == 0 {
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
	switch {
	case err == service.ErrProductNotFound:
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "product_not_found",
			Message: "Product not found",
		})
	case err == service.ErrInsufficientStock:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "insufficient_stock",
			Message: err.Error(),
		})
	case err == service.ErrInvalidQuantity:
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_quantity",
			Message: "Quantity must be at least 1",
		})
	default:
		log.Printf("Cart error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "An unexpected error occurred",
		})
	}
}
