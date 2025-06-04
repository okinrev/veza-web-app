//file: backend/routes/ressource.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-web-app/handlers"
	"veza-web-app/middleware"
)

func RegisterRessourceRoutes(r *mux.Router) {
	protected := r.PathPrefix("/products/{id}/docs").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("", handlers.ListInternalDocs).Methods("GET")
	protected.HandleFunc("/{id:[0-9]+}", handlers.GetInternalRessource).Methods("GET")
	r.HandleFunc("/docs/{id}", handlers.ServeInternalDocByID).Methods("GET")

}
