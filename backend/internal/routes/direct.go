// cmd/server/main.go
package main

import (
	"log"
	"net/http"
	"os"
	"time"

	// Internal packages
	"veza-web-app/internal/database"
	"veza-web-app/internal/routes"

	// External packages
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	Environment string
	Debug       bool
}

func loadConfig() Config {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found")
	}

	config := Config{
		DatabaseURL: os.Getenv("DATABASE_URL"),
		JWTSecret:   os.Getenv("JWT_SECRET"),
		Port:        os.Getenv("PORT"),
		Environment: os.Getenv("ENVIRONMENT"),
	}

	// Set defaults
	if config.Port == "" {
		config.Port = "8080"
	}
	if config.Environment == "" {
		config.Environment = "development"
	}
	config.Debug = config.Environment != "production"

	// Validate required config
	if config.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}
	if config.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	return config
}

func setupCORS(router *gin.Engine) {
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080", "http://localhost:5173"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With"},
		ExposeHeaders:    []string{"Content-Length", "Content-Disposition"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	router.Use(cors.New(corsConfig))
}

func setupFrontendRoutes(router *gin.Engine) {
	// Frontend HTML routes with clean URLs
	frontendRoutes := map[string]string{
		"/":                    "../frontend/public/main.html",
		"/chat":                "../frontend/public/chat.html",
		"/dashboard":           "../frontend/public/dashboard.html",
		"/dm":                  "../frontend/public/dm.html",
		"/gg":                  "../frontend/public/gg.html",
		"/hub":                 "../frontend/public/hub.html",
		"/hub_v2":              "../frontend/public/hub_v2.html",
		"/listings":            "../frontend/public/listings.html",
		"/login":               "../frontend/public/login.html",
		"/message":             "../frontend/public/message.html",
		"/plouf":               "../frontend/public/plouf.html",
		"/register":            "../frontend/public/register.html",
		"/room":                "../frontend/public/room.html",
		"/search":              "../frontend/public/search.html",
		"/search_v2":           "../frontend/public/search_v2.html",
		"/shared_ressources":   "../frontend/public/shared_ressources.html",
		"/test":                "../frontend/public/test.html",
		"/track":               "../frontend/public/track.html",
		"/user_products":       "../frontend/public/user_products.html",
		"/users":               "../frontend/public/users.html",
		"/api":                 "../frontend/public/api.html",
	}

	for path, file := range frontendRoutes {
		func(filePath string) {
			router.GET(path, func(c *gin.Context) {
				c.File(filePath)
			})
		}(file)
	}

	// Favicon
	router.GET("/favicon.ico", func(c *gin.Context) {
		c.File("../frontend/public/favicon.ico")
	})

	// Static asset directories
	router.StaticFS("/public", http.Dir("../frontend/public"))
	router.StaticFS("/admin", http.Dir("../frontend/admin"))
	router.StaticFS("/shared", http.Dir("../frontend/shared"))
	router.Static("/static", "./static")
}

func main() {
	// Load configuration
	config := loadConfig()

	// Set Gin mode
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	// Initialize database
	db, err := database.NewConnection(config.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Run database migrations
	if err := database.RunMigrations(db); err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// Initialize Gin router
	router := gin.New()

	// Global middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Setup CORS
	setupCORS(router)

	// Setup frontend routes
	setupFrontendRoutes(router)

	// Health check endpoint
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":     "ok",
			"service":    "veza-web-app",
			"version":    "1.0.0",
			"timestamp":  time.Now().Unix(),
			"environment": config.Environment,
		})
	})

	// Initialize API router with clean structure
	apiRouter := routes.NewRouter(routes.Config{
		DB:        db,
		JWTSecret: config.JWTSecret,
		Debug:     config.Debug,
	})

	// Setup all API routes
	apiRouter.SetupRoutes()

	// Mount API routes under /api/v1
	v1 := router.Group("/api/v1")
	{
		// Get the API engine and mount its routes
		apiEngine := apiRouter.GetEngine()
		
		// Mount all API routes from the clean router
		v1.Any("/*path", gin.WrapH(apiEngine))
	}

	// Legacy compatibility routes (direct paths without /api/v1 prefix)
	// These maintain compatibility with existing frontend code
	legacyRouter := routes.NewRouter(routes.Config{
		DB:        db,
		JWTSecret: config.JWTSecret,
		Debug:     config.Debug,
	})
	legacyRouter.SetupDirectRoutes()
	legacyEngine := legacyRouter.GetEngine()

	// Mount legacy routes directly
	router.NoRoute(func(c *gin.Context) {
		// Skip if it's already an API route or static file
		if len(c.Request.URL.Path) > 1 && 
		   !gin.IsDebugging() && 
		   c.Request.URL.Path[:5] != "/api/" && 
		   c.Request.URL.Path[:8] != "/public/" &&
		   c.Request.URL.Path[:7] != "/admin/" &&
		   c.Request.URL.Path[:8] != "/shared/" &&
		   c.Request.URL.Path[:8] != "/static/" {
			legacyEngine.ServeHTTP(c.Writer, c.Request)
		} else {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Route not found",
				"path":  c.Request.URL.Path,
			})
		}
	})

	// Configure server with timeouts
	server := &http.Server{
		Addr:         ":" + config.Port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Start server
	log.Printf("🚀 Talas server starting...")
	log.Printf("📍 Environment: %s", config.Environment)
	log.Printf("🔗 Health check: http://localhost:%s/health", config.Port)
	log.Printf("📖 API v1: http://localhost:%s/api/v1", config.Port)
	log.Printf("🌐 Frontend: http://localhost:%s", config.Port)
	log.Printf("📚 API Docs: http://localhost:%s/api", config.Port)
	
	if config.Debug {
		log.Printf("🐛 Debug mode enabled")
		log.Printf("📊 Database: Connected to %s", maskDatabaseURL(config.DatabaseURL))
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("❌ Failed to start server:", err)
	}
}

// maskDatabaseURL masks sensitive parts of the database URL for logging
func maskDatabaseURL(url string) string {
	if len(url) > 20 {
		return url[:10] + "***" + url[len(url)-7:]
	}
	return "***"
}