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
	"veza-web-app/internal/api/user"
	"veza-web-app/internal/api/track"
	"veza-web-app/internal/api/shared_ressources"
	"veza-web-app/internal/api/listing"
	"veza-web-app/internal/api/offer"
	"veza-web-app/internal/api/message"
	"veza-web-app/internal/api/room"
	"veza-web-app/internal/api/search"
	"veza-web-app/internal/api/tag"
	"veza-web-app/internal/api/suggestions"
	"veza-web-app/internal/admin/handlers"
	filesUtils "veza-web-app/internal/utils/files"

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
	if err := database.RunMigrations(db); err != nil {
	    log.Println("Failed to run migrations:", err)
	}

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
	router.GET("/api", func(c *gin.Context) {
		c.File("../frontend/public/api.html")
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

	// User service setup
	userService := user.NewService(db)
	userHandler := user.NewHandler(userService)

	// Protected user routes - AJOUTER CES LIGNES
	userGroup := v1.Group("/users")
	userGroup.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		// Profile management
		userGroup.GET("/me", userHandler.GetMe)                    // GET /api/v1/users/me
		userGroup.PUT("/me", userHandler.UpdateMe)                 // PUT /api/v1/users/me
		userGroup.PUT("/password", userHandler.ChangePassword)     // PUT /api/v1/users/password
		
		// User listing and search
		userGroup.GET("", userHandler.GetUsers)                    // GET /api/v1/users
		userGroup.GET("/except-me", userHandler.GetUsersExceptMe)  // GET /api/v1/users/except-me
		userGroup.GET("/search", userHandler.SearchUsers)          // GET /api/v1/users/search
		userGroup.GET("/:id/avatar", userHandler.GetUserAvatar)    // GET /api/v1/users/{id}/avatar
	}

	// Track routes - AJOUTER DANS main.go après userGroup
	trackHandler := track.NewHandler()
	trackGroup := v1.Group("/tracks")
	trackGroup.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		trackGroup.POST("", gin.WrapF(trackHandler.AddTrackWithUpload))    // POST /api/v1/tracks
		trackGroup.GET("", gin.WrapF(trackHandler.ListTracks))             // GET /api/v1/tracks (remplace /tracks/all)
		trackGroup.GET("/:id", gin.WrapF(trackHandler.GetTrack))           // GET /api/v1/tracks/{id}
		trackGroup.PUT("/:id", gin.WrapF(trackHandler.UpdateTrack))        // PUT /api/v1/tracks/{id}
		trackGroup.DELETE("/:id", gin.WrapF(trackHandler.DeleteTrack))     // DELETE /api/v1/tracks/{id}
	}

	// Streaming routes (publiques ou avec validation par signature)
	v1.GET("/generate-stream-url", middleware.JWTAuthMiddleware(jwtSecret), gin.WrapF(filesUtils.HandleGenerateSignedURL))
	router.GET("/stream/:filename", gin.WrapF(filesUtils.StreamAudioWithValidation))

	// Shared Resources routes - AJOUTER dans main.go
	sharedHandler := shared_ressources.NewHandler()
	sharedGroup := v1.Group("/shared_ressources")
	sharedGroup.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		sharedGroup.POST("", gin.WrapF(sharedHandler.UploadSharedResource))       // POST /api/v1/shared_ressources
		sharedGroup.GET("", gin.WrapF(sharedHandler.ListSharedResources))         // GET /api/v1/shared_ressources
		sharedGroup.GET("/search", gin.WrapF(sharedHandler.SearchSharedResources)) // GET /api/v1/shared_ressources/search
		sharedGroup.PUT("/:id", gin.WrapF(sharedHandler.UpdateSharedResource))    // PUT /api/v1/shared_ressources/{id}
		sharedGroup.DELETE("/:id", gin.WrapF(sharedHandler.DeleteSharedResource)) // DELETE /api/v1/shared_ressources/{id}
	}

	// Route publique pour télécharger les fichiers
	router.GET("/shared_ressources/:filename", gin.WrapF(sharedHandler.ServeSharedFile)) // GET /shared_ressources/{filename}

	// Listings routes - AJOUTER dans main.go
	listingHandler := listing.NewHandler()
	listingGroup := v1.Group("/listings")
	listingGroup.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		listingGroup.POST("", gin.WrapF(listingHandler.CreateListing))           // POST /api/v1/listings
		listingGroup.GET("", gin.WrapF(listingHandler.GetAllListings))          // GET /api/v1/listings
		listingGroup.GET("/:id", gin.WrapF(listingHandler.GetListingByID))      // GET /api/v1/listings/{id}
		listingGroup.DELETE("/:id", gin.WrapF(listingHandler.DeleteListing))    // DELETE /api/v1/listings/{id}
	}

	// Offers routes - AJOUTER dans main.go
	offerHandler := offer.NewHandler()
	v1.POST("/listings/:id/offer", middleware.JWTAuthMiddleware(jwtSecret), gin.WrapF(offerHandler.CreateOffer))        // POST /api/v1/listings/{id}/offer
	v1.POST("/offers/:id/accept", middleware.JWTAuthMiddleware(jwtSecret), gin.WrapF(offerHandler.AcceptOffer))         // POST /api/v1/offers/{id}/accept

	// Chat routes - AJOUTER dans main.go
	messageHandler := message.NewHandler()
	roomHandler := room.NewHandler()

	chatGroup := v1.Group("/chat")
	chatGroup.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		// Direct messages
		chatGroup.GET("/dm/:user_id", gin.WrapF(messageHandler.GetDmHandler))    // GET /api/v1/chat/dm/{user_id}
		
		// Rooms
		chatGroup.GET("/rooms", gin.WrapF(roomHandler.GetPublicRoomsHandler))               // GET /api/v1/chat/rooms
		chatGroup.POST("/rooms", gin.WrapF(roomHandler.CreateRoomHandler))                  // POST /api/v1/chat/rooms
		chatGroup.GET("/rooms/:room/messages", gin.WrapF(roomHandler.GetRoomMessagesHandler)) // GET /api/v1/chat/rooms/{room}/messages
	}

	// Search & Tags routes - AJOUTER dans main.go
	searchHandler := search.NewHandler()
	tagHandler := tag.NewHandler()
	suggestionHandler := suggestions.NewHandler()

	// Search routes
	v1.GET("/search", gin.WrapF(searchHandler.GlobalSearchHandler))                                          // GET /api/v1/search
	v1.GET("/search/advanced", middleware.JWTAuthMiddleware(jwtSecret), gin.WrapF(searchHandler.AdvancedSearchHandler)) // GET /api/v1/search/advanced
	v1.GET("/autocomplete", gin.WrapF(searchHandler.AutocompleteHandler))                                   // GET /api/v1/autocomplete

	// Tags routes
	tagGroup := v1.Group("/tags")
	{
		tagGroup.GET("", gin.WrapF(tagHandler.GetAllTags))           // GET /api/v1/tags
		tagGroup.GET("/search", gin.WrapF(tagHandler.SearchTags))    // GET /api/v1/tags/search
	}

	// Suggestions route
	v1.GET("/suggestions", gin.WrapF(suggestionHandler.GetSuggestions))  // GET /api/v1/suggestions

	// Admin routes - AJOUTER dans main.go après les autres groupes
	adminHandler := handlers.NewAdminHandler()
	adminGroup := router.Group("/admin-api")
	adminGroup.Use(middleware.JWTAuthMiddleware(jwtSecret))
	adminGroup.Use(middleware.AdminMiddleware())
	{
		// Dashboard
		adminGroup.GET("/dashboard", gin.WrapF(adminHandler.Dashboard))           // GET /admin-api/dashboard
		
		// Products (placeholders pour connecter vos services existants)
		adminGroup.GET("/products", gin.WrapF(adminHandler.ListProducts))        // GET /admin-api/products
		adminGroup.POST("/products", gin.WrapF(adminHandler.CreateProduct))      // POST /admin-api/products
		
		// Categories
		adminGroup.GET("/categories", gin.WrapF(adminHandler.ListCategories))    // GET /admin-api/categories
		
		// TODO: Ajouter d'autres routes admin selon vos besoins
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