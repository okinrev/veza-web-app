// internal/routes/listings.go
package routes

import (
	"github.com/gin-gonic/gin"
	"veza-web-app/internal/handlers"
	"veza-web-app/internal/middleware"
)

func (r *Router) setupListingRoutes(rg *gin.RouterGroup) {
	listingHandler := handlers.NewListingOfferHandler(r.db)

	listings := rg.Group("/listings")
	{
		// Public listing endpoints
		listings.GET("", listingHandler.GetAllListings)
		listings.GET("/:id", listingHandler.GetListingByID)

		// Protected listing endpoints
		protected := listings.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			protected.POST("", listingHandler.CreateListing)
			protected.PUT("/:id", listingHandler.UpdateListing)
			protected.DELETE("/:id", listingHandler.DeleteListing)
			protected.GET("/:id/offers", listingHandler.GetListingOffers)
		}
	}
}

func (r *Router) setupOfferRoutes(rg *gin.RouterGroup) {
	offerHandler := handlers.NewOfferHandler(r.db)
	listingHandler := handlers.NewListingOfferHandler(r.db) // For legacy routes

	offers := rg.Group("/offers")
	offers.Use(middleware.JWTAuthMiddleware(r.jwtSecret)) // All offer routes require auth
	{
		// Offer management
		offers.GET("", offerHandler.GetUserOffers)
		offers.GET("/stats", offerHandler.GetOfferStats)
		offers.GET("/:id", offerHandler.GetOffer)
		offers.PUT("/:id", offerHandler.UpdateOffer)
		
		// Offer actions
		offers.POST("/listings/:listing_id", offerHandler.CreateOffer)
	}

	// Legacy routes for compatibility
	legacy := rg.Group("/")
	legacy.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
	{
		// Legacy offer creation
		legacy.POST("/listings/:id/offer", listingHandler.CreateOffer)
		legacy.POST("/offers/:id/accept", listingHandler.AcceptOffer)
	}
}