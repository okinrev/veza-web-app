// internal/routes/router.go
package routes

import (
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"veza-web-app/internal/database"
	"veza-web-app/internal/handlers"
	"veza-web-app/internal/middleware"
)

type Router struct {
	db        *database.DB
	jwtSecret string
	engine    *gin.Engine
}

type Config struct {
	DB        *database.DB
	JWTSecret string
	Debug     bool
}

func NewRouter(config Config) *Router {
	// Set Gin mode
	if !config.Debug {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Global middleware
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// CORS configuration
	corsConfig := cors.Config{
		AllowOrigins:     []string{"http://localhost:3000", "http://localhost:8080"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}
	engine.Use(cors.New(corsConfig))

	return &Router{
		db:        config.DB,
		jwtSecret: config.JWTSecret,
		engine:    engine,
	}
}

func (r *Router) SetupRoutes() {
	// Health check
	r.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ok",
			"timestamp": time.Now().Unix(),
		})
	})

	// API version 1
	v1 := r.engine.Group("/api/v1")
	{
		r.setupAuthRoutes(v1)
		r.setupUserRoutes(v1)
		r.setupTrackRoutes(v1)
		r.setupSharedResourceRoutes(v1)
		r.setupProductRoutes(v1)
		r.setupUserProductRoutes(v1)
		r.setupFileRoutes(v1)
		r.setupChatRoutes(v1)
		r.setupRoomRoutes(v1)
		r.setupListingRoutes(v1)
		r.setupOfferRoutes(v1)
		r.setupSearchRoutes(v1)
		r.setupTagRoutes(v1)
		r.setupAdminRoutes(v1)
	}

	// Direct routes (maintain compatibility)
	r.setupDirectRoutes()
}

func (r *Router) GetEngine() *gin.Engine {
	return r.engine
}

// SetupDirectRoutes sets up legacy compatibility routes
func (r *Router) SetupDirectRoutes() {
	r.setupDirectRoutes()
}

// setupDirectRoutes maintains compatibility with existing frontend
func (r *Router) setupDirectRoutes() {
	// Initialize handlers
	authHandler := handlers.NewAuthHandler(r.db, r.jwtSecret)
	userHandler := handlers.NewUserHandler(r.db)
	trackHandler := handlers.NewTrackHandler(r.db)
	sharedResourcesHandler := handlers.NewSharedResourcesHandler(r.db)
	tagsSearchHandler := handlers.NewTagsSearchHandler(r.db)
	fileHandler := handlers.NewFileHandler(r.db)

	// Authentication routes (public)
	r.engine.POST("/signup", authHandler.Register)
	r.engine.POST("/login", authHandler.Login)
	r.engine.POST("/refresh", authHandler.RefreshToken)
	r.engine.POST("/logout", authHandler.Logout)

	// User routes
	authRequired := r.engine.Group("/")
	authRequired.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
	{
		authRequired.GET("/users/me", authHandler.GetMe)
		authRequired.PUT("/users/me", userHandler.UpdateMe)
		authRequired.PUT("/users/password", userHandler.ChangePassword)
		authRequired.GET("/users", userHandler.GetUsers)
		authRequired.GET("/users/search", userHandler.SearchUsers)
		authRequired.GET("/users/except-me", userHandler.GetUsersExceptMe)
		authRequired.GET("/users/:id", userHandler.GetUserByID)
		authRequired.GET("/users/:id/avatar", userHandler.GetUserAvatar)
		authRequired.POST("/users/avatar", userHandler.UploadAvatar)
		authRequired.DELETE("/users/avatar", userHandler.DeleteAvatar)
	}

	// Track routes
	r.engine.GET("/tracks", trackHandler.ListTracks)
	r.engine.GET("/tracks/:id", trackHandler.GetTrack)
	authRequired.POST("/tracks", trackHandler.AddTrackWithUpload)
	authRequired.PUT("/tracks/:id", trackHandler.UpdateTrack)
	authRequired.DELETE("/tracks/:id", trackHandler.DeleteTrack)
	authRequired.GET("/tracks/:id/stats", trackHandler.GetTrackStats)

	// Streaming routes
	r.engine.GET("/stream/:filename", trackHandler.StreamAudio)
	r.engine.GET("/stream/signed/:filename", trackHandler.StreamAudioSigned)
	authRequired.GET("/generate-stream-url", trackHandler.GenerateStreamURL)

	// Shared resources
	r.engine.GET("/shared_resources", sharedResourcesHandler.ListSharedResources)
	r.engine.GET("/shared_resources/search", sharedResourcesHandler.SearchSharedResources)
	r.engine.GET("/shared_resources/types", sharedResourcesHandler.GetPredefinedResourceTypes)
	r.engine.GET("/shared_resources/stats", sharedResourcesHandler.GetDetailedStats)
	r.engine.GET("/shared_resources/:filename", sharedResourcesHandler.ServeSharedFile)
	authRequired.POST("/shared_resources", sharedResourcesHandler.UploadSharedResource)
	authRequired.PUT("/shared_resources/:id", sharedResourcesHandler.UpdateSharedResource)
	authRequired.DELETE("/shared_resources/:id", sharedResourcesHandler.DeleteSharedResource)
	authRequired.GET("/shared_resources/:id/stats", sharedResourcesHandler.GetDownloadStats)

	// Tags and search
	r.engine.GET("/tags", tagsSearchHandler.GetAllTags)
	r.engine.GET("/tags/search", tagsSearchHandler.SearchTags)
	r.engine.GET("/tags/trending", tagsSearchHandler.GetTrendingTags)
	r.engine.GET("/search", tagsSearchHandler.GlobalSearch)
	r.engine.GET("/autocomplete", tagsSearchHandler.GetAutocomplete)
	authRequired.GET("/search/advanced", tagsSearchHandler.AdvancedSearch)
	authRequired.GET("/suggestions", tagsSearchHandler.GetSuggestions)
	authRequired.GET("/suggestions/contextual", tagsSearchHandler.GetContextualSuggestions)

	// File serving routes
	r.engine.GET("/files/:filename", fileHandler.ServeInternalDoc)
	r.engine.GET("/avatars/:filename", userHandler.ServeAvatar)
}