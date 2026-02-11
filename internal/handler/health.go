package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type HealthHandler struct{}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Check if the server is running
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func (h *HealthHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Server is running",
	})
}

// Welcome godoc
// @Summary Welcome endpoint
// @Description Welcome message for the API
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Router / [get]
func (h *HealthHandler) Welcome(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "Welcome to Foods & Drinks API",
		"version": "v1",
	})
}
