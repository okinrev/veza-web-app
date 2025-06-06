package auth

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	auth := router.Group("/auth")
	{
		// Routes publiques
		auth.POST("/register", handler.Register)
		auth.POST("/signup", handler.Register) // Alias
		auth.POST("/login", handler.Login)
		auth.POST("/refresh", handler.RefreshToken)
		auth.POST("/logout", handler.Logout)

		// Routes protégées
		protected := auth.Group("")
		protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
		{
			protected.GET("/me", handler.GetMe)
		}
	}
}
