// internal/routes/shared_resources.go
package routes

import (
	"github.com/gin-gonic/gin"
	"veza-web-app/internal/handlers"
	"veza-web-app/internal/middleware"
)

func (r *Router) setupSharedResourceRoutes(rg *gin.RouterGroup) {
	resourceHandler := handlers.NewSharedResourcesHandler(r.db)

	resources := rg.Group("/shared-resources")
	{
		// Public resource endpoints
		resources.GET("", resourceHandler.ListSharedResources)
		resources.GET("/search", resourceHandler.SearchSharedResources)
		resources.GET("/types", resourceHandler.GetPredefinedResourceTypes)
		resources.GET("/stats", resourceHandler.GetDetailedStats)
		resources.GET("/:id", resourceHandler.GetDetailedStats) // Get specific resource
		
		// File serving
		resources.GET("/files/:filename", resourceHandler.ServeSharedFile)

		// Protected resource endpoints
		protected := resources.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			protected.POST("", resourceHandler.UploadSharedResource)
			protected.PUT("/:id", resourceHandler.UpdateSharedResource)
			protected.DELETE("/:id", resourceHandler.DeleteSharedResource)
			protected.GET("/:id/stats", resourceHandler.GetDownloadStats)
		}
	}

	// Legacy compatibility routes
	legacy := rg.Group("/shared_resources")
	{
		legacy.GET("", resourceHandler.ListSharedResources)
		legacy.GET("/search", resourceHandler.SearchSharedResources)
		legacy.GET("/types", resourceHandler.GetPredefinedResourceTypes)
		legacy.GET("/stats", resourceHandler.GetDetailedStats)
		legacy.GET("/:filename", resourceHandler.ServeSharedFile)

		protected := legacy.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			protected.POST("", resourceHandler.UploadSharedResource)
			protected.PUT("/:id", resourceHandler.UpdateSharedResource)
			protected.DELETE("/:id", resourceHandler.DeleteSharedResource)
			protected.GET("/:id/stats", resourceHandler.GetDownloadStats)
		}
	}
}