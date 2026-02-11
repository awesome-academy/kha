package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kha/foods-drinks/internal/handler"
)

func SetupRouter(healthHandler *handler.HealthHandler) *gin.Engine {
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check (public)
	router.GET("/health", healthHandler.HealthCheck)
	router.GET("/", healthHandler.Welcome)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("")
		{
			_ = public // TODO: Add public routes (login, register, products list, etc.)
		}

		// Protected routes (require authentication)
		protected := v1.Group("")
		{
			_ = protected // TODO: Add auth middleware and protected routes
		}

		// Admin routes (require admin role)
		admin := v1.Group("/admin")
		{
			_ = admin // TODO: Add admin middleware and admin routes
		}
	}

	return router
}
