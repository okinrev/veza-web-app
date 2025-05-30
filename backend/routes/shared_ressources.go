//file: backend/routes/shared_ressources.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-backend/handlers"
	"veza-backend/middleware"
)

func RegisterSharedRoutes(r *mux.Router) {
	protected := r.PathPrefix("/shared_ressources").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("", handlers.UploadSharedResource).Methods("POST")
	protected.HandleFunc("", handlers.ListSharedResources).Methods("GET")
	r.HandleFunc("/shared_ressources/search", handlers.SearchSharedResources).Methods("GET")
	r.HandleFunc("/shared_ressources/{filename}", handlers.ServeSharedFile).Methods("GET")

}
