// internal/routes/products.go
package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/okinrev/veza-web-app/internal/handlers"
	"github.com/okinrev/veza-web-app/internal/middleware"
)

func (r *Router) setupProductRoutes(rg *gin.RouterGroup) {
	productHandler := handlers.NewProductHandler(r.db)

	products := rg.Group("/products")
	{
		// Public product endpoints
		products.GET("/search", productHandler.SearchProducts)

		// Admin-only product management
		protected := products.Group("/")
		protected.Use(middleware.JWTAuthMiddleware(r.jwtSecret))
		{
			protected.GET("", productHandler.GetProducts)
			protected.GET("/:id", productHandler.GetProduct)
			protected.POST("", productHandler.CreateProduct)
			protected.PUT("/:id", productHandler.UpdateProduct)
			protected.DELETE("/:id", productHandler.DeleteProduct)
		}
	}
}

func (r *Router) setupUserProductRoutes(rg *gin.RouterGroup) {
	userProductHandler := handlers.NewUserProductHandler(r.db)

	userProducts := rg.Group("/user-products")
	userProducts.Use(middleware.JWTAuthMiddleware(r.jwtSecret)) // All require auth
	{
		// User product collection management
		userProducts.GET("", userProductHandler.ListUserProducts)
		userProducts.GET("/search", userProductHandler.SearchUserProducts)
		userProducts.GET("/warranty", userProductHandler.GetWarrantyStatus)
		userProducts.GET("/:id", userProductHandler.GetUserProduct)
		userProducts.POST("", userProductHandler.CreateUserProduct)
		userProducts.PUT("/:id", userProductHandler.UpdateUserProduct)
		userProducts.DELETE("/:id", userProductHandler.DeleteUserProduct)
	}
}

func (r *Router) setupFileRoutes(rg *gin.RouterGroup) {
	fileHandler := handlers.NewFileHandler(r.db)

	files := rg.Group("/files")
	files.Use(middleware.JWTAuthMiddleware(r.jwtSecret)) // All require auth
	{
		// File management for user products
		files.GET("/products/:id", fileHandler.ListProductFiles)
		files.POST("/products/:id", fileHandler.UploadFile)
		files.GET("/:id", fileHandler.DownloadFile)
		files.DELETE("/:id", fileHandler.DeleteFile)
		files.GET("/:id/stats", fileHandler.GetFileStats)

		// Internal documents
		docs := files.Group("/docs")
		{
			docs.GET("/products/:id", fileHandler.ListInternalDocs)
			docs.POST("/products/:id", fileHandler.UploadInternalDoc)
			docs.GET("/:id", fileHandler.ServeInternalDoc)
		}
	}

	// Legacy file serving routes
	rg.GET("/docs/:id", fileHandler.ServeInternalDoc)
}