//file: backend/routes/auth.go

package routes

import (
	"veza-backend/middleware"
	"github.com/gorilla/mux"
	"veza-backend/handlers"
)

func RegisterAuthRoutes(r *mux.Router) {
	r.HandleFunc("/signup", handlers.SignupHandler).Methods("POST")
	r.HandleFunc("/login", handlers.LoginHandler).Methods("POST")
	r.HandleFunc("/refresh", handlers.RefreshHandler).Methods("POST")


	protected := r.PathPrefix("/").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)
	protected.HandleFunc("/me", handlers.MeHandler).Methods("GET")
	protected.HandleFunc("/users/password", handlers.ChangePasswordHandler).Methods("PUT")
	protected.HandleFunc("/users/{id}", handlers.UpdateUserHandler).Methods("PUT")
	protected.HandleFunc("/users/{id}", handlers.DeleteUserHandler).Methods("DELETE")
}
