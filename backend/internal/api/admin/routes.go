package admin

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func SetupRoutes(router *gin.RouterGroup, handler *Handler, jwtSecret string) {
	admin := router.Group("/admin")
	admin.Use(middleware.JWTAuthMiddleware(jwtSecret))
	admin.Use(middleware.AdminMiddleware())
	{
		admin.GET("/dashboard", handler.Dashboard)
		admin.GET("/users", handler.GetUsers)
		admin.GET("/analytics", handler.GetAnalytics)
		admin.GET("/categories", handler.GetCategories)
	}
}
