package main

import (
    "log"
    "net/http"
    "os"

    // Internal packages
    "veza-web-app/internal/config"
    "veza-web-app/internal/database"
    
    // API packages
    "veza-web-app/internal/api/auth"
    "veza-web-app/internal/api/user"
    "veza-web-app/internal/api/middleware"
    
    // Utility packages
    "veza-web-app/pkg/logger"

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
    cfg := config.New()

    // Initialize logger
    logger := logger.New(cfg.Environment)

    // Initialize database
    db, err := database.NewConnection(cfg.DatabaseURL)
    if err != nil {
        log.Fatal("Failed to connect to database:", err)
    }
    defer db.Close()

    // Run migrations
    if err := database.RunMigrations(db); err != nil {
        log.Fatal("Failed to run migrations:", err)
    }

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
        authService := auth.NewService(db, cfg.JWTSecret)
        auth.SetupRoutes(v1, authService)
        
        // User routes
        userService := user.NewService(db)
        user.SetupRoutes(v1, userService)
    }

    // Admin routes with JWT middleware
    admin := router.Group("/admin")
    admin.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
    admin.Use(middleware.AdminMiddleware())
    {
        // Add admin routes here
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