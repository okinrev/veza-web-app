// file: backend/routes/tag.go

package services

import (
	"github.com/gorilla/mux"
	"veza-web-app/internal/handlers"
)

func RegisterTagRoutes(r *mux.Router) {
	tags := r.PathPrefix("/tags").Subrouter()
	tags.HandleFunc("", handlers.GetAllTags).Methods("GET")        // GET /tags
	tags.HandleFunc("/search", handlers.SearchTags).Methods("GET") // GET /tags/search?q=hip
}
