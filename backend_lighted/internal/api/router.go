package api

import (
	"github.com/gin-gonic/gin"
	
	"github.com/okinrev/veza-web-app/internal/config"
	"github.com/okinrev/veza-web-app/internal/database"
	
	"github.com/okinrev/veza-web-app/internal/api/auth"
	"github.com/okinrev/veza-web-app/internal/api/user"
	"github.com/okinrev/veza-web-app/internal/api/admin"
	"github.com/okinrev/veza-web-app/internal/api/track"
	"github.com/okinrev/veza-web-app/internal/api/listing"
	"github.com/okinrev/veza-web-app/internal/api/offer"
	"github.com/okinrev/veza-web-app/internal/api/message"
	"github.com/okinrev/veza-web-app/internal/api/room"
	"github.com/okinrev/veza-web-app/internal/api/search"
	"github.com/okinrev/veza-web-app/internal/api/tag"
	"github.com/okinrev/veza-web-app/internal/api/shared_resources"
)

// SetupRoutes configure toutes les routes API
func SetupRoutes(router *gin.Engine, db *database.DB, cfg *config.Config) {
	// Groupe API v1
	v1 := router.Group("/api/v1")
	
	// Auth routes
	authService := auth.NewService(db, cfg.JWT.Secret)
	authHandler := auth.NewHandler(authService)
	auth.SetupRoutes(v1, authHandler, cfg.JWT.Secret)
	
	// User routes
	userService := user.NewService(db)
	userHandler := user.NewHandler(userService)
	user.SetupRoutes(v1, userHandler, cfg.JWT.Secret)
	
	// Admin routes
	adminService := admin.NewService(db)
	adminHandler := admin.NewHandler(adminService)
	admin.SetupRoutes(v1, adminHandler, cfg.JWT.Secret)
	
	// Track routes
	trackService := track.NewService(db, cfg.JWT.Secret)
	trackHandler := track.NewHandler(trackService)
	track.SetupRoutes(v1, trackHandler, cfg.JWT.Secret)
	
	// Listing routes
	listingService := listing.NewService(db)
	listingHandler := listing.NewHandler(listingService)
	listing.SetupRoutes(v1, listingHandler, cfg.JWT.Secret)
	
	// Offer routes
	offerService := offer.NewService(db)
	offerHandler := offer.NewHandler(offerService)
	offer.SetupRoutes(v1, offerHandler, cfg.JWT.Secret)
	
	// Message routes
	messageService := message.NewService(db)
	messageHandler := message.NewHandler(messageService)
	message.SetupRoutes(v1, messageHandler, cfg.JWT.Secret)
	
	// Room routes
	roomService := room.NewService(db)
	roomHandler := room.NewHandler(roomService)
	room.SetupRoutes(v1, roomHandler, cfg.JWT.Secret)
	
	// Search routes
	searchService := search.NewService(db)
	searchHandler := search.NewHandler(searchService)
	search.SetupRoutes(v1, searchHandler, cfg.JWT.Secret)
	
	// Tag routes
	tagService := tag.NewService(db)
	tagHandler := tag.NewHandler(tagService)
	tag.SetupRoutes(v1, tagHandler, cfg.JWT.Secret)
	
	// Shared resources routes
	sharedResourcesService := shared_resources.NewService(db)
	sharedResourcesHandler := shared_resources.NewHandler(sharedResourcesService)
	shared_resources.SetupRoutes(v1, sharedResourcesHandler, cfg.JWT.Secret)
}