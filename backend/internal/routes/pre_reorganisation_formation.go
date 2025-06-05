// file: routes/formation.go

 package routes

// import (
// 	"github.com/gorilla/mux"
// 	"veza-backend/handlers"
// 	"veza-backend/middleware"
// )

// func RegisterFormationRoutes(r *mux.Router) {
// 	protected := r.PathPrefix("/formations").Subrouter()
// 	protected.Use(middleware.JWTAuthMiddleware)

// 	protected.HandleFunc("", handlers.CreateFormation).Methods("POST")
// 	protected.HandleFunc("", handlers.ListFormations).Methods("GET")
// 	protected.HandleFunc("/{id}", handlers.GetFormation).Methods("GET")
// 	protected.HandleFunc("/{id}/progress", handlers.UpdateProgress).Methods("POST")
// }
