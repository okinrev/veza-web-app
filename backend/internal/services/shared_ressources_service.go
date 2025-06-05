//file: backend/routes/shared_ressources.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-web-app/handlers"
	"veza-web-app/middleware"
)

func RegisterSharedRoutes(r *mux.Router) {
	protected := r.PathPrefix("/shared_ressources").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("", handlers.UploadSharedResource).Methods("POST")
	protected.HandleFunc("", handlers.ListSharedResources).Methods("GET")
	protected.HandleFunc("/search", handlers.SearchSharedResources).Methods("GET")
	protected.HandleFunc("/{id:[0-9]+}", handlers.UpdateSharedResource).Methods("PUT")
	protected.HandleFunc("/{id:[0-9]+}", handlers.DeleteSharedResource).Methods("DELETE")

	// Visualisation publique
	r.HandleFunc("/shared_ressources/{filename}", handlers.ServeSharedFile).Methods("GET")
}
