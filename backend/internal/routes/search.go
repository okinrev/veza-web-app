// internal/routes/search.go
package routes

import (
	"github.com/gin-gonic/gin"
	"veza-web-app/internal/handlers"
	"veza-web-app/internal/middleware"
)

func (r *Router) setupSearchRoutes(rg *gin.RouterGroup) {
	searchHandler := handlers.NewTagsSearchHandler(r.db)

	search := rg.Group("/search")
	{
		// Public search endpoints
		search.GET("", searchHandler.GlobalSearch)
		search.GET("/autocomplete", searchHandler.GetAutocomplete)

		// Protected search endpoints
		protected := search.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			protected.GET("/advanced", searchHandler.AdvancedSearch)
			protected.GET("/suggestions", searchHandler.GetSuggestions)
			protected.GET("/contextual", searchHandler.GetContextualSuggestions)
		}
	}

	// Legacy autocomplete route
	rg.GET("/autocomplete", searchHandler.GetAutocomplete)
}

func (r *Router) setupTagRoutes(rg *gin.RouterGroup) {
	tagHandler := handlers.NewTagsSearchHandler(r.db)

	tags := rg.Group("/tags")
	{
		// Public tag endpoints
		tags.GET("", tagHandler.GetAllTags)
		tags.GET("/search", tagHandler.SearchTags)
		tags.GET("/trending", tagHandler.GetTrendingTags)

		// Protected tag endpoints
		protected := tags.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			// Admin-only cache management
			protected.DELETE("/cache", tagHandler.ClearSuggestionCache)
		}
	}

	// Legacy suggestions route
	rg.GET("/suggestions", tagHandler.GetSuggestions)
}