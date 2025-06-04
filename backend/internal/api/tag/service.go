// file: backend/routes/tag.go

package routes

import (
	"github.com/gorilla/mux"
	"backend/handlers"
)

func RegisterTagRoutes(r *mux.Router) {
	tags := r.PathPrefix("/tags").Subrouter()
	tags.HandleFunc("", handlers.GetAllTags).Methods("GET")        // GET /tags
	tags.HandleFunc("/search", handlers.SearchTags).Methods("GET") // GET /tags/search?q=hip
}
