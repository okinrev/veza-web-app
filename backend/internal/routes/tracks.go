// internal/routes/tracks.go
package routes

import (
	"github.com/gin-gonic/gin"
	"veza-web-app/internal/handlers"
	"veza-web-app/internal/middleware"
)

func (r *Router) setupTrackRoutes(rg *gin.RouterGroup) {
	trackHandler := handlers.NewTrackHandler(r.db)

	tracks := rg.Group("/tracks")
	{
		// Public track endpoints
		tracks.GET("", trackHandler.ListTracks)
		tracks.GET("/:id", trackHandler.GetTrack)
		tracks.GET("/:id/stats", trackHandler.GetTrackStats)

		// Protected track endpoints
		protected := tracks.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			protected.POST("", trackHandler.AddTrackWithUpload)
			protected.PUT("/:id", trackHandler.UpdateTrack)
			protected.DELETE("/:id", trackHandler.DeleteTrack)
		}
	}

	// Streaming endpoints (separate group for different middleware)
	streaming := rg.Group("/stream")
	{
		// Public streaming
		streaming.GET("/:filename", trackHandler.StreamAudio)
		streaming.GET("/signed/:filename", trackHandler.StreamAudioSigned)

		// Protected streaming utilities
		protected := streaming.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			protected.GET("/generate-url", trackHandler.GenerateStreamURL)
		}
	}
}