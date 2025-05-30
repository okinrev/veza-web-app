//file: backend/routes/products.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-backend/handlers"
	"veza-backend/middleware"
)

func RegisterProductRoutes(r *mux.Router) {
	protected := r.PathPrefix("/products").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("", handlers.ListUserProducts).Methods("GET")
	protected.HandleFunc("", handlers.CreateProduct).Methods("POST")
	protected.HandleFunc("/{id:[0-9]+}", handlers.GetProductDetails).Methods("GET")
	protected.HandleFunc("/{id:[0-9]+}", handlers.UpdateProduct).Methods("PUT")
	protected.HandleFunc("/{id:[0-9]+}", handlers.DeleteProduct).Methods("DELETE")
}
