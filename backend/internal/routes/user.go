// internal/routes/users.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/handlers"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func (r *Router) setupUserRoutes(rg *gin.RouterGroup) {
	userHandler := handlers.NewUserHandler(r.db)

	users := rg.Group("/users")
	{
		// Public user endpoints
		users.GET("", userHandler.GetUsers)
		users.GET("/search", userHandler.SearchUsers)
		users.GET("/:id", userHandler.GetUserByID)
		users.GET("/:id/avatar", userHandler.GetUserAvatar)

		// File serving
		users.GET("/avatars/:filename", userHandler.ServeAvatar)

		// Protected user endpoints
		protected := users.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			// Profile management
			protected.GET("/me", userHandler.GetUserByID) // Will get from context
			protected.PUT("/me", userHandler.UpdateMe)
			protected.PUT("/password", userHandler.ChangePassword)

			// Avatar management
			protected.POST("/avatar", userHandler.UploadAvatar)
			protected.DELETE("/avatar", userHandler.DeleteAvatar)

			// User discovery
			protected.GET("/except-me", userHandler.GetUsersExceptMe)
		}
	}
}