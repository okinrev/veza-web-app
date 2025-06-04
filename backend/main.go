package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"veza-web-app/internal/config"
	"veza-web-app/internal/database"
	
	// Admin handlers
	adminHandlers "veza-web-app/internal/admin/handlers"
	adminServices "veza-web-app/internal/admin/services"
	
	// API handlers
	authHandler "veza-web-app/internal/api/auth"
	exchangeHandler "veza-web-app/internal/api/exchange"
	fileHandler "veza-web-app/internal/api/file"
	formationHandler "veza-web-app/internal/api/formation"
	listingHandler "veza-web-app/internal/api/listing"
	messageHandler "veza-web-app/internal/api/message"
	offerHandler "veza-web-app/internal/api/offer"
	productsHandler "veza-web-app/internal/api/products"
	ressourceHandler "veza-web-app/internal/api/ressource"
	roomHandler "veza-web-app/internal/api/room"
	searchHandler "veza-web-app/internal/api/search"
	sharedRessourcesHandler "veza-web-app/internal/api/shared_ressources"
	suggestionsHandler "veza-web-app/internal/api/suggestions"
	tagHandler "veza-web-app/internal/api/tag"
	trackHandler "veza-web-app/internal/api/track"
	userHandler "veza-web-app/internal/api/user"
	userProductsHandler "veza-web-app/internal/api/user_products"
	
	// Middleware
	"veza-web-app/internal/api/middleware"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize configuration
	cfg := config.New()

	// Initialize database connection
	db, err := database.NewConnection(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run database migrations
	if err := database.RunMigrations(db); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Initialize Gin router
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	
	router := gin.Default()

	// Configure CORS
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = []string{"http://localhost:3000", "http://localhost:8080"}
	corsConfig.AllowCredentials = true
	corsConfig.AllowHeaders = []string{"Origin", "Content-Length", "Content-Type", "Authorization"}
	router.Use(cors.New(corsConfig))

	// Serve static files
	router.Static("/static", "./static")
	router.Static("/uploads", "./static/uploads")
	router.Static("/shared", "./static/shared")
	router.Static("/shared_ressources", "./static/shared_ressources")

	// Initialize services
	adminProductService := adminServices.NewProductService(db)

	authService := authHandler.NewService(db, cfg.JWTSecret)
	exchangeService := exchangeHandler.NewService(db)
	fileService := fileHandler.NewService(db)
	formationService := formationHandler.NewService(db)
	listingService := listingHandler.NewService(db)
	messageService := messageHandler.NewService(db)
	offerService := offerHandler.NewService(db)
	productsService := productsHandler.NewService(db)
	ressourceService := ressourceHandler.NewService(db)
	roomService := roomHandler.NewService(db)
	searchService := searchHandler.NewService(db)
	sharedRessourcesService := sharedRessourcesHandler.NewService(db)
	suggestionsService := suggestionsHandler.NewService(db)
	tagService := tagHandler.NewService(db)
	trackService := trackHandler.NewService(db)
	userService := userHandler.NewService(db)
	userProductsService := userProductsHandler.NewService(db)

	// Initialize handlers
	adminAnalyticsHandler := adminHandlers.NewAnalyticsHandler()
	adminCategoryHandler := adminHandlers.NewCategoryHandler()
	adminProductsHandler := adminHandlers.NewProductsHandler(adminProductService)

	fileHandlerInstance := fileHandler.NewHandler(fileService)
	listingHandlerInstance := listingHandler.NewHandler(listingService)
	messageHandlerInstance := messageHandler.NewHandler(messageService)
	offerHandlerInstance := offerHandler.NewHandler(offerService)
	productsHandlerInstance := productsHandler.NewHandler(productsService)
	ressourceHandlerInstance := ressourceHandler.NewHandler(ressourceService)
	roomHandlerInstance := roomHandler.NewHandler(roomService)
	searchHandlerInstance := searchHandler.NewHandler(searchService)
	sharedRessourcesHandlerInstance := sharedRessourcesHandler.NewHandler(sharedRessourcesService)
	suggestionsHandlerInstance := suggestionsHandler.NewHandler(suggestionsService)
	tagHandlerInstance := tagHandler.NewHandler(tagService)
	trackHandlerInstance := trackHandler.NewHandler(trackService)
	userHandlerInstance := userHandler.NewHandler(userService)
	userProductsHandlerInstance := userProductsHandler.NewHandler(userProductsService)

	// API Routes
	api := router.Group("/api")
	
	// Auth routes (no middleware)
	api.POST("/login", authService.Login)
	api.POST("/register", authService.Register)
	api.POST("/refresh", authService.RefreshToken)

	// Protected API routes
	protected := api.Group("/")
	protected.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
	{
		// User routes
		protected.GET("/users", userHandlerInstance.GetUsers)
		protected.GET("/users/:id", userHandlerInstance.GetUser)
		protected.PUT("/users/:id", userHandlerInstance.UpdateUser)
		protected.DELETE("/users/:id", userHandlerInstance.DeleteUser)

		// Product routes
		protected.GET("/products", productsHandlerInstance.GetProducts)
		protected.POST("/products", productsHandlerInstance.CreateProduct)
		protected.GET("/products/:id", productsHandlerInstance.GetProduct)
		protected.PUT("/products/:id", productsHandlerInstance.UpdateProduct)
		protected.DELETE("/products/:id", productsHandlerInstance.DeleteProduct)

		// User products routes
		protected.GET("/user-products", userProductsHandlerInstance.GetUserProducts)
		protected.POST("/user-products", userProductsHandlerInstance.CreateUserProduct)
		protected.GET("/user-products/:id", userProductsHandlerInstance.GetUserProduct)
		protected.PUT("/user-products/:id", userProductsHandlerInstance.UpdateUserProduct)
		protected.DELETE("/user-products/:id", userProductsHandlerInstance.DeleteUserProduct)

		// File routes
		protected.POST("/files/upload", fileHandlerInstance.UploadFile)
		protected.GET("/files/:id", fileHandlerInstance.GetFile)
		protected.DELETE("/files/:id", fileHandlerInstance.DeleteFile)

		// Listing routes
		protected.GET("/listings", listingHandlerInstance.GetListings)
		protected.POST("/listings", listingHandlerInstance.CreateListing)
		protected.GET("/listings/:id", listingHandlerInstance.GetListing)
		protected.PUT("/listings/:id", listingHandlerInstance.UpdateListing)
		protected.DELETE("/listings/:id", listingHandlerInstance.DeleteListing)

		// Message routes
		protected.GET("/messages", messageHandlerInstance.GetMessages)
		protected.POST("/messages", messageHandlerInstance.CreateMessage)
		protected.GET("/messages/:id", messageHandlerInstance.GetMessage)
		protected.PUT("/messages/:id", messageHandlerInstance.UpdateMessage)
		protected.DELETE("/messages/:id", messageHandlerInstance.DeleteMessage)

		// Room routes
		protected.GET("/rooms", roomHandlerInstance.GetRooms)
		protected.POST("/rooms", roomHandlerInstance.CreateRoom)
		protected.GET("/rooms/:id", roomHandlerInstance.GetRoom)
		protected.PUT("/rooms/:id", roomHandlerInstance.UpdateRoom)
		protected.DELETE("/rooms/:id", roomHandlerInstance.DeleteRoom)

		// Offer routes
		protected.GET("/offers", offerHandlerInstance.GetOffers)
		protected.POST("/offers", offerHandlerInstance.CreateOffer)
		protected.GET("/offers/:id", offerHandlerInstance.GetOffer)
		protected.PUT("/offers/:id", offerHandlerInstance.UpdateOffer)
		protected.DELETE("/offers/:id", offerHandlerInstance.DeleteOffer)

		// Resource routes
		protected.GET("/ressources", ressourceHandlerInstance.GetRessources)
		protected.POST("/ressources", ressourceHandlerInstance.CreateRessource)
		protected.GET("/ressources/:id", ressourceHandlerInstance.GetRessource)
		protected.PUT("/ressources/:id", ressourceHandlerInstance.UpdateRessource)
		protected.DELETE("/ressources/:id", ressourceHandlerInstance.DeleteRessource)

		// Shared resources routes
		protected.GET("/shared-ressources", sharedRessourcesHandlerInstance.GetSharedRessources)
		protected.POST("/shared-ressources", sharedRessourcesHandlerInstance.CreateSharedRessource)
		protected.GET("/shared-ressources/:id", sharedRessourcesHandlerInstance.GetSharedRessource)
		protected.PUT("/shared-ressources/:id", sharedRessourcesHandlerInstance.UpdateSharedRessource)
		protected.DELETE("/shared-ressources/:id", sharedRessourcesHandlerInstance.DeleteSharedRessource)

		// Search routes
		protected.GET("/search", searchHandlerInstance.Search)

		// Suggestions routes
		protected.GET("/suggestions", suggestionsHandlerInstance.GetSuggestions)
		protected.POST("/suggestions", suggestionsHandlerInstance.CreateSuggestion)

		// Tag routes
		protected.GET("/tags", tagHandlerInstance.GetTags)
		protected.POST("/tags", tagHandlerInstance.CreateTag)
		protected.GET("/tags/:id", tagHandlerInstance.GetTag)
		protected.PUT("/tags/:id", tagHandlerInstance.UpdateTag)
		protected.DELETE("/tags/:id", tagHandlerInstance.DeleteTag)

		// Track routes
		protected.GET("/tracks", trackHandlerInstance.GetTracks)
		protected.POST("/tracks", trackHandlerInstance.CreateTrack)
		protected.GET("/tracks/:id", trackHandlerInstance.GetTrack)
		protected.PUT("/tracks/:id", trackHandlerInstance.UpdateTrack)
		protected.DELETE("/tracks/:id", trackHandlerInstance.DeleteTrack)

		// Exchange routes
		protected.GET("/exchanges", exchangeService.GetExchanges)
		protected.POST("/exchanges", exchangeService.CreateExchange)

		// Formation routes
		protected.GET("/formations", formationService.GetFormations)
		protected.POST("/formations", formationService.CreateFormation)
	}

	// Admin routes
	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware(cfg.JWTSecret))
	// TODO: Add admin role middleware
	{
		// Analytics
		admin.GET("/analytics", adminAnalyticsHandler.GetAnalytics)
		admin.GET("/analytics/users", adminAnalyticsHandler.GetUserAnalytics)
		admin.GET("/analytics/products", adminAnalyticsHandler.GetProductAnalytics)

		// Categories
		admin.GET("/categories", adminCategoryHandler.GetCategories)
		admin.POST("/categories", adminCategoryHandler.CreateCategory)
		admin.GET("/categories/:id", adminCategoryHandler.GetCategory)
		admin.PUT("/categories/:id", adminCategoryHandler.UpdateCategory)
		admin.DELETE("/categories/:id", adminCategoryHandler.DeleteCategory)

		// Admin products
		admin.GET("/products", adminProductsHandler.GetProducts)
		admin.POST("/products", adminProductsHandler.CreateProduct)
		admin.GET("/products/:id", adminProductsHandler.GetProduct)
		admin.PUT("/products/:id", adminProductsHandler.UpdateProduct)
		admin.DELETE("/products/:id", adminProductsHandler.DeleteProduct)
	}

	// Frontend routes (serve HTML pages)
	frontend := router.Group("/")
	{
		frontend.Static("/assets", "./frontend/public/assets")
		frontend.LoadHTMLGlob("frontend/public/*.html")
		
		frontend.GET("/", func(c *gin.Context) {
			c.HTML(http.StatusOK, "main.html", gin.H{"title": "Veza - Home"})
		})
		frontend.GET("/login", func(c *gin.Context) {
			c.HTML(http.StatusOK, "login.html", gin.H{"title": "Veza - Login"})
		})
		frontend.GET("/register", func(c *gin.Context) {
			c.HTML(http.StatusOK, "register.html", gin.H{"title": "Veza - Register"})
		})
		frontend.GET("/dashboard", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dashboard.html", gin.H{"title": "Veza - Dashboard"})
		})
		frontend.GET("/chat", func(c *gin.Context) {
			c.HTML(http.StatusOK, "chat.html", gin.H{"title": "Veza - Chat"})
		})
		frontend.GET("/hub", func(c *gin.Context) {
			c.HTML(http.StatusOK, "hub.html", gin.H{"title": "Veza - Hub"})
		})
		frontend.GET("/hub_v2", func(c *gin.Context) {
			c.HTML(http.StatusOK, "hub_v2.html", gin.H{"title": "Veza - Hub V2"})
		})
		frontend.GET("/room", func(c *gin.Context) {
			c.HTML(http.StatusOK, "room.html", gin.H{"title": "Veza - Room"})
		})
		frontend.GET("/dm", func(c *gin.Context) {
			c.HTML(http.StatusOK, "dm.html", gin.H{"title": "Veza - Direct Messages"})
		})
		frontend.GET("/listings", func(c *gin.Context) {
			c.HTML(http.StatusOK, "listings.html", gin.H{"title": "Veza - Listings"})
		})
		frontend.GET("/search", func(c *gin.Context) {
			c.HTML(http.StatusOK, "search.html", gin.H{"title": "Veza - Search"})
		})
		frontend.GET("/search_v2", func(c *gin.Context) {
			c.HTML(http.StatusOK, "search_v2.html", gin.H{"title": "Veza - Search V2"})
		})
		frontend.GET("/shared_ressources", func(c *gin.Context) {
			c.HTML(http.StatusOK, "shared_ressources.html", gin.H{"title": "Veza - Shared Resources"})
		})
		frontend.GET("/track", func(c *gin.Context) {
			c.HTML(http.StatusOK, "track.html", gin.H{"title": "Veza - Track"})
		})
		frontend.GET("/user_products", func(c *gin.Context) {
			c.HTML(http.StatusOK, "user_products.html", gin.H{"title": "Veza - User Products"})
		})
		frontend.GET("/users", func(c *gin.Context) {
			c.HTML(http.StatusOK, "users.html", gin.H{"title": "Veza - Users"})
		})
		frontend.GET("/message", func(c *gin.Context) {
			c.HTML(http.StatusOK, "message.html", gin.H{"title": "Veza - Messages"})
		})
		frontend.GET("/test", func(c *gin.Context) {
			c.HTML(http.StatusOK, "test.html", gin.H{"title": "Veza - Test"})
		})
		frontend.GET("/plouf", func(c *gin.Context) {
			c.HTML(http.StatusOK, "plouf.html", gin.H{"title": "Veza - Plouf"})
		})
		frontend.GET("/gg", func(c *gin.Context) {
			c.HTML(http.StatusOK, "gg.html", gin.H{"title": "Veza - GG"})
		})
	}

	// Admin frontend routes
	adminFrontend := router.Group("/admin")
	{
		adminFrontend.Static("/assets", "./frontend/admin/assets")
		adminFrontend.LoadHTMLGlob("frontend/admin/*.html")
		
		adminFrontend.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusMovedPermanently, "/admin/products")
		})
		adminFrontend.GET("/products", func(c *gin.Context) {
			c.HTML(http.StatusOK, "admin_products.html", gin.H{"title": "Veza Admin - Products"})
		})
		adminFrontend.GET("/api", func(c *gin.Context) {
			c.HTML(http.StatusOK, "api.html", gin.H{"title": "Veza Admin - API Documentation"})
		})
	}

	// Create HTTP server
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", cfg.Port),
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("🚀 Server starting on port %s", cfg.Port)
		log.Printf("🌐 Frontend: http://localhost:%s", cfg.Port)
		log.Printf("📡 API: http://localhost:%s/api", cfg.Port)
		log.Printf("⚡ Admin: http://localhost:%s/admin", cfg.Port)
		
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Give outstanding requests a deadline for completion
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("✅ Server exiting")
}