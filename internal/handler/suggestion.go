package handler

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/kha/foods-drinks/internal/dto"
	"github.com/kha/foods-drinks/internal/middleware"
	"github.com/kha/foods-drinks/internal/service"
)

type SuggestionHandler struct {
	suggestionService *service.SuggestionService
}

func NewSuggestionHandler(suggestionService *service.SuggestionService) *SuggestionHandler {
	return &SuggestionHandler{suggestionService: suggestionService}
}

// Create godoc
// @Summary Create product suggestion
// @Description Authenticated user creates a new food/drink suggestion
// @Tags suggestions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateSuggestionRequest true "Create suggestion request"
// @Success 201 {object} dto.SuggestionResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /api/v1/suggestions [post]
func (h *SuggestionHandler) Create(c *gin.Context) {
	userID, ok := middleware.GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, dto.ErrorResponse{Error: "unauthorized", Message: "Authentication required"})
		return
	}

	var req dto.CreateSuggestionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "validation_error", Message: "Invalid request: " + err.Error()})
		return
	}

	resp, err := h.suggestionService.Create(userID, &req)
	if err != nil {
		h.handleSuggestionError(c, err)
		return
	}

	c.JSON(http.StatusCreated, resp)
}

func (h *SuggestionHandler) handleSuggestionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrCategoryNotFound):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "category_not_found", Message: "Category not found"})
	case errors.Is(err, service.ErrInvalidSuggestionState):
		c.JSON(http.StatusBadRequest, dto.ErrorResponse{Error: "invalid_suggestion_state", Message: "Invalid suggestion state"})
	default:
		log.Printf("Suggestion error: %v", err)
		c.JSON(http.StatusInternalServerError, dto.ErrorResponse{Error: "internal_error", Message: "An unexpected error occurred"})
	}
}
