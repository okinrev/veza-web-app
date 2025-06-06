package track

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	tracks := router.Group("/tracks")
	
	// Routes publiques
	tracks.GET("", handler.ListTracks)
	tracks.GET("/:id", handler.GetTrack)
	
	// Routes protégées
	protected := tracks.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		protected.POST("", handler.AddTrackWithUpload)
		protected.PUT("/:id", handler.UpdateTrack)
		protected.DELETE("/:id", handler.DeleteTrack)
	}
}