// file: backend/routes/user.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-backend/handlers"
	"veza-backend/middleware"
)

func RegisterUserRoutes(r *mux.Router) {
	// Routes protégées sous /users
	protected := r.PathPrefix("/users").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("", handlers.GetAllUsers).Methods("GET")                 // /users
	protected.HandleFunc("/search", handlers.SearchUsers).Methods("GET")         // /users/search?q=...
	protected.HandleFunc("/except-me", handlers.GetUsersExceptMe).Methods("GET") // /users/except-me
	protected.HandleFunc("/{id:[0-9]+}", handlers.GetUserByID).Methods("GET")    // /users/{id}
	protected.HandleFunc("/{id:[0-9]+}/avatar", handlers.GetUserAvatar).Methods("GET") // /users/{id}/avatar
}
