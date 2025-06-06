// internal/routes/auth.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/handlers"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func (r *Router) setupAuthRoutes(rg *gin.RouterGroup) {
	authHandler := handlers.NewAuthHandler(r.db, r.jwtSecret)

	auth := rg.Group("/auth")
	{
		// Public authentication endpoints
		auth.POST("/register", authHandler.Register)
		auth.POST("/signup", authHandler.Register) // Alias for compatibility
		auth.POST("/login", authHandler.Login)
		auth.POST("/refresh", authHandler.RefreshToken)
		auth.POST("/logout", authHandler.Logout)

		// Protected authentication endpoints
		protected := auth.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			protected.GET("/me", authHandler.GetMe)
			protected.PUT("/me", authHandler.GetMe) // For profile updates via auth
		}
	}
}