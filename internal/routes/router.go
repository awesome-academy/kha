package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/kha/foods-drinks/internal/handler"
)

// RouterDependencies holds all dependencies for router setup
type RouterDependencies struct {
	HealthHandler  *handler.HealthHandler
	AuthHandler    *handler.AuthHandler
	CorsMiddleware gin.HandlerFunc
}

func SetupRouter(deps *RouterDependencies) *gin.Engine {
	router := gin.New()

	// Global middleware - order matters!
	router.Use(deps.CorsMiddleware) // CORS first
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Health check (public)
	router.GET("/health", deps.HealthHandler.HealthCheck)
	router.GET("/", deps.HealthHandler.Welcome)

	// API v1 routes
	v1 := router.Group("/api/v1")
	{
		// Auth routes (public)
		auth := v1.Group("/auth")
		{
			auth.POST("/register", deps.AuthHandler.Register)
			auth.POST("/login", deps.AuthHandler.Login)
		}

		// Public routes
		public := v1.Group("")
		{
			_ = public // TODO: Add public routes (products list, etc.)
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
