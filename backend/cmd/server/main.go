// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"veza-web-app/internal/config"
	"veza-web-app/internal/database"
	"veza-web-app/internal/routes"

	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Load configuration
	cfg := config.New()

	// Initialize database
	db, err := database.NewConnection(cfg.Database.URL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run database migrations
	if err := database.RunMigrations(db); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// Initialize router
	router := routes.NewRouter(routes.Config{
		DB:        db,
		JWTSecret: cfg.JWT.Secret,
		Debug:     cfg.Server.Environment == "development",
	})

	// Setup routes
	router.SetupRoutes()

	// Configure server
	server := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      router.GetEngine(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	log.Printf("🚀 Talas server starting on port %s...", cfg.Server.Port)
	log.Printf("📍 Environment: %s", cfg.Server.Environment)
	log.Printf("🔗 Health check: http://localhost:%s/health", cfg.Server.Port)
	log.Printf("📖 API v1: http://localhost:%s/api/v1", cfg.Server.Port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ Failed to start server:", err)
	}
}