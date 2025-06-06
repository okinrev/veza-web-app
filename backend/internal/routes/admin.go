// internal/routes/admin.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/handlers"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func (r *Router) setupAdminRoutes(rg *gin.RouterGroup) {
	adminHandler := handlers.NewAdminHandler(r.db)

	admin := rg.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware(r.jwtSecret)) // All admin routes require auth
	{
		// Dashboard and analytics
		admin.GET("/dashboard", adminHandler.Dashboard)
		admin.GET("/analytics", adminHandler.GetAnalytics)
		admin.GET("/users", adminHandler.GetUsers)

		// Category management
		categories := admin.Group("/categories")
		{
			categories.GET("", adminHandler.GetCategories)
			categories.POST("", adminHandler.CreateCategory)
			categories.PUT("/:id", adminHandler.UpdateCategory)
			categories.DELETE("/:id", adminHandler.DeleteCategory)
		}

		// System management
		system := admin.Group("/system")
		{
			// Cache management
			system.DELETE("/cache/suggestions", func(c *gin.Context) {
				tagHandler := handlers.NewTagsSearchHandler(r.db)
				tagHandler.ClearSuggestionCache(c)
			})
		}
	}
}