// file: backend/routes/suggestions.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-backend/handlers"
)

func RegisterSuggestionRoutes(r *mux.Router) {
	r.HandleFunc("/suggestions", handlers.GetSuggestions).Methods("GET")
}
