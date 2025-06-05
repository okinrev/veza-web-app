// file: backend/routes/suggestions.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-web-app/handlers"
)

func RegisterSuggestionRoutes(r *mux.Router) {
	r.HandleFunc("/suggestions", handlers.GetSuggestions).Methods("GET")
}
