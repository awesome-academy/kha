package handler

import (
	"errors"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/service"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.RegisterRequest true "Register request"
// @Success 201 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 409 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Trim whitespace
	req.Email = strings.TrimSpace(req.Email)
	req.FullName = strings.TrimSpace(req.FullName)

	resp, err := h.authService.Register(&req)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

// Login godoc
// @Summary Login user
// @Description Login with email and password
// @Tags auth
// @Accept json
// @Produce json
// @Param request body dto.LoginRequest true "Login request"
// @Success 200 {object} dto.AuthResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	var req dto.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.handleValidationError(c, err)
		return
	}

	// Trim whitespace
	req.Email = strings.TrimSpace(req.Email)

	resp, err := h.authService.Login(&req)
	if err != nil {
		h.handleAuthError(c, err)
		return
	}

	c.JSON(http.StatusOK, resp)
}

// handleValidationError handles validation errors
func (h *AuthHandler) handleValidationError(c *gin.Context, err error) {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		details := make(map[string]string)
		for _, fe := range ve {
			field := strings.ToLower(fe.Field())
			switch fe.Tag() {
			case "required":
				details[field] = field + " is required"
			case "email":
				details[field] = "invalid email format"
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
		Error:   "bad_request",
		Message: "Invalid request body",
	})
}

// handleAuthError handles authentication errors
func (h *AuthHandler) handleAuthError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrEmailAlreadyExists):
		c.JSON(http.StatusConflict, dto.ErrorResponse{
			Error:   "email_exists",
			Message: "Email already exists",
		})
	case errors.Is(err, service.ErrInvalidCredentials):
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{
			Error:   "invalid_credentials",
			Message: "Invalid email or password",
		})
	case errors.Is(err, service.ErrUserInactive):
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "user_inactive",
			Message: "Your account is inactive",
		})
	case errors.Is(err, service.ErrUserBanned):
		c.JSON(http.StatusForbidden, dto.ErrorResponse{
			Error:   "user_banned",
			Message: "Your account has been banned",
		})
	default:
		log.Printf("Internal error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{
			Error:   "internal_error",
			Message: "An unexpected error occurred",
		})
	}
}
