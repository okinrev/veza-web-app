//file: backend/routes/shared_ressources.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-backend/handlers"
	"veza-backend/middleware"
)

func RegisterSharedRoutes(r *mux.Router) {
	protected := r.PathPrefix("/shared").Subrouter()
	protected.Use(middleware.JWTAuthMiddleware)

	protected.HandleFunc("", handlers.UploadSharedResource).Methods("POST")
	protected.HandleFunc("", handlers.ListSharedResources).Methods("GET")
	r.HandleFunc("/shared/{filename}", handlers.ServeSharedFile).Methods("GET")
}
