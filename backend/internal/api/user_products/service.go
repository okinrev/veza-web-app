//file: backend/routes/user_products.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-backend/handlers"
	"veza-backend/middleware"
)

func RegisterUserProductRoutes(r *mux.Router) {
	// Toutes les routes user-products nécessitent une authentification
	protected := r.PathPrefix("/user-products").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("", handlers.ListUserProducts).Methods("GET")
	protected.HandleFunc("", handlers.CreateUserProduct).Methods("POST")
	protected.HandleFunc("/{id:[0-9]+}", handlers.GetUserProductDetails).Methods("GET")
	protected.HandleFunc("/{id:[0-9]+}", handlers.UpdateUserProduct).Methods("PUT")
	protected.HandleFunc("/{id:[0-9]+}", handlers.DeleteUserProduct).Methods("DELETE")
}