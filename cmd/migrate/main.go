package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/kha/foods-drinks/internal/config"
)

func main() {
	configPath := flag.String("config", "config.yaml", "path to config file")
	migrationsPath := flag.String("migrations", "migrations", "path to migrations folder")
	command := flag.String("command", "up", "migration command: up, down, version, force")
	steps := flag.Int("steps", 0, "number of migrations to run (0 = all)")
	forceVersion := flag.Int("force", -1, "force migration version")
	flag.Parse()

	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("mysql://%s:%s@tcp(%s:%d)/%s?multiStatements=true",
		cfg.Database.Username,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.DBName,
	)

	m, err := migrate.New(
		fmt.Sprintf("file://%s", *migrationsPath),
		dsn,
	)
	if err != nil {
		log.Fatalf("Failed to create migrate instance: %v", err)
	}
	defer func() {
		sourceErr, dbErr := m.Close()
		if sourceErr != nil {
			log.Printf("Warning: failed to close migration source: %v", sourceErr)
		}
		if dbErr != nil {
			log.Printf("Warning: failed to close database connection: %v", dbErr)
		}
	}()

	switch *command {
	case "up":
		if *steps > 0 {
			err = m.Steps(*steps)
		} else {
			err = m.Up()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migration up completed successfully")

	case "down":
		if *steps > 0 {
			err = m.Steps(-*steps)
		} else {
			err = m.Down()
		}
		if err != nil && err != migrate.ErrNoChange {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migration down completed successfully")

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("Current version: %d, Dirty: %v\n", version, dirty)

	case "force":
		if *forceVersion < 0 {
			log.Fatal("Please provide a version to force with -force flag")
		}
		err = m.Force(*forceVersion)
		if err != nil {
			log.Fatalf("Force version failed: %v", err)
		}
		fmt.Printf("Forced version to %d\n", *forceVersion)

	default:
		fmt.Println("Available commands: up, down, version, force")
		os.Exit(1)
	}
}
