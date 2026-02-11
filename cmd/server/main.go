package main

import (
	"fmt"
	"log"

	"github.com/kha/foods-drinks/internal/config"
	"github.com/kha/foods-drinks/pkg/database"
)

func main() {
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	db, err := database.NewConnection(&cfg.Database)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	sqlDB, _ := db.DB()
	defer sqlDB.Close()

	fmt.Printf("Server %s starting on port %d...\n", cfg.App.Name, cfg.App.Port)
	fmt.Println("Database connected successfully!")

	// TODO: Initialize routes and start HTTP server
}
