package main

import (
    "log"
    "net/http"
    "os"

    // Internal packages
    "veza-web-app/internal/config"
    "veza-web-app/internal/database"
    
    // Admin packages
    adminHandlers "veza-web-app/internal/admin/handlers"
    adminServices "veza-web-app/internal/admin/services"
    
    // API packages
    "veza-web-app/internal/api/auth"
    "veza-web-app/internal/api/exchange"
    "veza-web-app/internal/api/file"
    "veza-web-app/internal/api/formation"
    "veza-web-app/internal/api/listing"
    "veza-web-app/internal/api/message"
    "veza-web-app/internal/api/middleware"
    "veza-web-app/internal/api/offer"
    "veza-web-app/internal/api/products"
    "veza-web-app/internal/api/ressource"
    "veza-web-app/internal/api/room"
    "veza-web-app/internal/api/search"
    "veza-web-app/internal/api/shared_ressources"
    "veza-web-app/internal/api/suggestions"
    "veza-web-app/internal/api/tag"
    "veza-web-app/internal/api/track"
    "veza-web-app/internal/api/user"
    "veza-web-app/internal/api/user_products"
    
    // Utility packages
    "veza-web-app/pkg/logger"
    "veza-web-app/pkg/validator"

    // External packages
    "github.com/gin-gonic/gin"
    "github.com/joho/godotenv"
)

func main() {
    // Load environment variables
    if err := godotenv.Load(); err != nil {
        log.Println("No .env file found")
    }

    // Initialize config
    cfg := config.Load()

    // Initialize logger
    logger.Init(cfg.LogLevel)

    // Initialize database
    db, err := database.NewConnection(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Initialize Gin router
    if cfg.Environment == "production" {
        gin.SetMode(gin.ReleaseMode)
    }

    router := gin.Default()

    // Add middleware
    router.Use(middleware.CORS())
    router.Use(middleware.Logger())

    // Health check endpoint
    router.GET("/health", func(c *gin.Context) {
        c.JSON(http.StatusOK, gin.H{
            "status": "ok",
            "service": "veza-web-app",
        })
    })

    // Setup API routes
    v1 := router.Group("/api/v1")
    {
        // Authentication routes
        auth.SetupRoutes(v1, db)
        
        // User routes
        user.SetupRoutes(v1, db)
        
        // Product routes
        products.SetupRoutes(v1, db)
        
        // File routes
        file.SetupRoutes(v1, db)
        
        // Chat routes
        message.SetupRoutes(v1, db)
        room.SetupRoutes(v1, db)
        
        // Trading routes
        listing.SetupRoutes(v1, db)
        offer.SetupRoutes(v1, db)
        exchange.SetupRoutes(v1, db)
        
        // Resource routes
        ressource.SetupRoutes(v1, db)
        shared_ressources.SetupRoutes(v1, db)
        
        // Media routes
        track.SetupRoutes(v1, db)
        
        // Search routes
        search.SetupRoutes(v1, db)
        suggestions.SetupRoutes(v1, db)
        tag.SetupRoutes(v1, db)
        
        // Formation routes
        formation.SetupRoutes(v1, db)
        
        // User products
        user_products.SetupRoutes(v1, db)
    }

    // Admin routes
    admin := router.Group("/admin")
    {
        adminHandlers.SetupRoutes(admin, db)
    }

    // Static file serving
    router.Static("/static", "./static")
    router.StaticFile("/favicon.ico", "./static/favicon.ico")

    // Start server
    port := cfg.Port
    if port == "" {
        port = "8080"
    }

    log.Printf("🚀 Server starting on port %s", port)
    log.Printf("📍 Health check: http://localhost:%s/health", port)
    log.Printf("📖 API docs: http://localhost:%s/api/v1", port)
    
    if err := router.Run(":" + port); err != nil {
        log.Fatal("❌ Failed to start server:", err)
    }
}