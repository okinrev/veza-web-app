package main

import (
	"log"
	"net/http"
	"os"

	// Internal packages
	"veza-web-app/internal/database"

	// API packages
	"veza-web-app/internal/api/auth"
	"veza-web-app/internal/api/middleware"

	// External packages
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	// Get database URL from environment
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	// Get JWT secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Initialize database
	db, err := database.NewConnection(databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run migrations
	//if err := database.RunMigrations(db); err != nil {
	//    log.Fatal("Failed to run migrations:", err)
	//}

	// Initialize Gin router
	environment := os.Getenv("ENVIRONMENT")
	if environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Add basic middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// --- Serve Frontend HTML Files with Clean URLs ---
	// Map desired URLs to specific HTML files using c.File()

	router.GET("/", func(c *gin.Context) {
		c.File("../frontend/public/main.html")
	})
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File("../frontend/public/favicon.ico")
	})
	router.GET("/chat", func(c *gin.Context) {
		c.File("../frontend/public/chat.html")
	})
	router.GET("/dashboard", func(c *gin.Context) {
		c.File("../frontend/public/dashboard.html")
	})
	router.GET("/dm", func(c *gin.Context) {
		c.File("../frontend/public/dm.html")
	})
	router.GET("/gg", func(c *gin.Context) {
		c.File("../frontend/public/gg.html")
	})
	router.GET("/hub", func(c *gin.Context) {
		c.File("../frontend/public/hub.html")
	})
	router.GET("/hub_v2", func(c *gin.Context) {
		c.File("../frontend/public/hub_v2.html")
	})
	router.GET("/listings", func(c *gin.Context) {
		c.File("../frontend/public/listings.html")
	})
	router.GET("/login", func(c *gin.Context) {
		c.File("../frontend/public/login.html")
	})
	router.GET("/message", func(c *gin.Context) {
		c.File("../frontend/public/message.html")
	})
	router.GET("/plouf", func(c *gin.Context) {
		c.File("../frontend/public/plouf.html")
	})
	router.GET("/register", func(c *gin.Context) {
		c.File("../frontend/public/register.html")
	})
	router.GET("/room", func(c *gin.Context) {
		c.File("../frontend/public/room.html")
	})
	router.GET("/search", func(c *gin.Context) {
		c.File("../frontend/public/search.html")
	})
	router.GET("/search_v2", func(c *gin.Context) {
		c.File("../frontend/public/search_v2.html")
	})
	router.GET("/shared_ressources", func(c *gin.Context) {
		c.File("../frontend/public/shared_ressources.html")
	})
	router.GET("/test", func(c *gin.Context) {
		c.File("../frontend/public/test.html")
	})
	router.GET("/track", func(c *gin.Context) {
		c.File("../frontend/public/track.html")
	})
	router.GET("/user_products", func(c *gin.Context) {
		c.File("../frontend/public/user_products.html")
	})
	router.GET("/users", func(c *gin.Context) {
		c.File("../frontend/public/users.html")
	})

	// Serve asset directories (CSS, JS, images, etc.) under their original paths
	// These remain the same because your HTML files will reference them with the /public/ prefix
	router.StaticFS("/public", http.Dir("../frontend/public"))
	router.StaticFS("/admin", http.Dir("../frontend/admin"))
	router.StaticFS("/shared", http.Dir("../frontend/shared"))

	// --- End Frontend Static File Serving ---

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "veza-web-app",
		})
	})

	// Setup API routes
	v1 := router.Group("/api/v1")
	{
		// Authentication routes
		authService := auth.NewService(db, jwtSecret)
		v1.POST("/login", authService.Login)
		v1.POST("/register", authService.Register)
		v1.POST("/refresh", authService.RefreshToken)
		v1.POST("/logout", authService.Logout)

		// Protected user routes
		userGroup := v1.Group("/users")
		userGroup.Use(middleware.JWTAuthMiddleware(jwtSecret))
		{
			userGroup.GET("", func(c *gin.Context) {
				// Implementation for getting users with pagination
				c.JSON(http.StatusOK, gin.H{"message": "Users endpoint"})
			})
			userGroup.GET("/:id", func(c *gin.Context) {
				// Implementation for getting user by ID
				c.JSON(http.StatusOK, gin.H{"message": "User by ID endpoint"})
			})
		}
	}

	// Admin routes with JWT and admin middleware
	admin := router.Group("/admin-api")
	admin.Use(middleware.JWTAuthMiddleware(jwtSecret))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Admin dashboard"})
		})
	}

	// Static file serving (for other static assets like general images, etc.)
	router.Static("/static", "./static")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📍 Health check: http://localhost:%s/health", port)
	log.Printf("📖 API: http://localhost:%s/api/v1", port)

	if err := router.Run(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}