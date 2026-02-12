package main

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/kha/foods-drinks/internal/config"
	"github.com/kha/foods-drinks/internal/handler"
	"github.com/kha/foods-drinks/internal/middleware"
	"github.com/kha/foods-drinks/internal/routes"
	"github.com/kha/foods-drinks/pkg/database"
)

func main() {
	// Load config
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Set Gin mode based on environment
	if cfg.App.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Connect to database
	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	log.Println("Database connected successfully!")

	// Initialize handlers
	healthHandler := handler.NewHealthHandler()

	// Setup router with CORS middleware
	router := routes.SetupRouter(healthHandler, middleware.CORSConfig())

	// Start server
	addr := fmt.Sprintf(":%d", cfg.App.Port)
	log.Printf("Server %s starting on %s", cfg.App.Name, addr)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
