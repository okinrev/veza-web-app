package user

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	users := router.Group("/users")
	
	// Routes publiques
	users.GET("", handler.GetUsers)
	users.GET("/:id/avatar", handler.GetUserAvatar)
	
	// Routes protégées
	protected := users.Group("")
	protected.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		protected.GET("/me", handler.GetMe)
		protected.PUT("/me", handler.UpdateMe)
		protected.PUT("/me/password", handler.ChangePassword)
		protected.GET("/except-me", handler.GetUsersExceptMe)
		protected.GET("/search", handler.SearchUsers)
	}
}