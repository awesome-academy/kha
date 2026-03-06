package handler

import (
	"errors"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/middleware"
	"github.com/kha/foods-drinks/internal/service"
)

var shippingPhonePattern = regexp.MustCompile(`^[0-9+\-()\s]+$`)

type OrderHandler struct {
	orderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
	return &OrderHandler{orderService: orderService}
}

// Create godoc
// @Summary Create order from cart
// @Description Create a new order from current user cart, snapshot item price/name, clear cart, and update stock
// @Tags orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateOrderRequest true "Create order request"
// @Success 201 {object} dto.OrderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/orders [post]
func (h *OrderHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req dto.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleCreateOrderValidationError(c, err)
		return
	}

	req.ShippingAddress = strings.TrimSpace(req.ShippingAddress)
	req.ShippingPhone = strings.TrimSpace(req.ShippingPhone)

	if details := validateCreateOrderRequest(&req); len(details) > 0 {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Validation failed",
			Details: details,
		})
		return
	}

	resp, err := h.orderService.CreateOrderFromCart(userID, &req)
	if err != nil {
		h.handleOrderError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// List godoc
// @Summary List current user orders
// @Description Get order history of current user with status/date filters and pagination
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param status query string false "pending|confirmed|processing|shipping|delivered|cancelled"
// @Param from_date query string false "From date (YYYY-MM-DD)"
// @Param to_date query string false "To date (YYYY-MM-DD)"
// @Success 200 {object} dto.PaginatedResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/orders [get]
func (h *OrderHandler) List(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	var req dto.OrderListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_params",
			Message: "Invalid query parameters: " + err.Error(),
		})
		return
	}

	resp, err := h.orderService.ListOrders(userID, &req)
	if err != nil {
		h.handleOrderError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// GetDetail godoc
// @Summary Get order detail
// @Description Get detail of one order for current user including items
// @Tags orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/orders/{id} [get]
func (h *OrderHandler) GetDetail(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "unauthorized",
			Message: "Authentication required",
		})
		return
	}

	orderID, valid := parsePositiveUint64(c.Param("id"))
	if !valid {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_order_id",
			Message: "Invalid order ID",
		})
		return
	}

	resp, err := h.orderService.GetOrderDetail(userID, uint(orderID))
	if err != nil {
		h.handleOrderError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) handleOrderError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCartEmpty):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "cart_empty",
			Message: "Cart is empty",
		})
	case errors.Is(err, service.ErrProductNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "product_not_found",
			Message: "Product not found",
		})
	case errors.Is(err, service.ErrInsufficientStock):
		message := "Insufficient stock"
		if err != nil && err.Error() != "" {
			message = err.Error()
		}
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "insufficient_stock",
			Message: message,
		})
	case errors.Is(err, service.ErrOrderNotFound):
		c.JSON(http.StatusNotFound, dto.ErrorResponse{
			Error:   "order_not_found",
			Message: "Order not found",
		})
	case errors.Is(err, service.ErrInvalidDateFilter):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_date_filter",
			Message: "Invalid date filter. Expected YYYY-MM-DD and from_date <= to_date",
		})
	case errors.Is(err, service.ErrInvalidOrderInput):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "invalid_order_input",
			Message: "Shipping address and shipping phone are required",
		})
	default:
		log.Printf("Order error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "An unexpected error occurred",
		})
	}
}

func parsePositiveUint64(raw string) (uint64, bool) {
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

func (h *OrderHandler) handleCreateOrderValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := map[string]string{}
		for _, fe := range ve {
			field := strings.ToLower(fe.Field())
			switch field {
			case "shippingaddress":
				field = "shipping_address"
			case "shippingphone":
				field = "shipping_phone"
			case "notes":
				field = "notes"
			}

			switch fe.Tag() {
			case "required":
				details[field] = field + " is required"
			case "min":
				details[field] = field + " must be at least " + fe.Param() + " characters"
			case "max":
				details[field] = field + " must be at most " + fe.Param() + " characters"
			default:
				details[field] = field + " is invalid"
			}
		}

		c.JSON(http.StatusBadRequest, dto.ErrorResponse{
			Error:   "validation_error",
			Message: "Validation failed",
			Details: details,
		})
		return
	}

	c.JSON(http.StatusBadRequest, dto.ErrorResponse{
		Error:   "validation_error",
		Message: "Invalid request body",
	})
}

func validateCreateOrderRequest(req *dto.CreateOrderRequest) map[string]string {
	details := map[string]string{}

	if strings.TrimSpace(req.ShippingAddress) == "" {
		details["shipping_address"] = "shipping_address is required"
	}

	phone := strings.TrimSpace(req.ShippingPhone)
	if phone == "" {
		details["shipping_phone"] = "shipping_phone is required"
		return details
	}

	if !shippingPhonePattern.MatchString(phone) {
		details["shipping_phone"] = "shipping_phone contains invalid characters"
		return details
	}

	digitCount := 0
	for i := 0; i < len(phone); i++ {
		if phone[i] >= '0' && phone[i] <= '9' {
			digitCount++
		}
	}
	if digitCount < 8 {
		details["shipping_phone"] = "shipping_phone must include at least 8 digits"
	}

	return details
}
