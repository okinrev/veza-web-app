//file: backend/routes/search.go
package routes

import (
	"github.com/gorilla/mux"
	"veza-backend/handlers"
	"github.com/jmoiron/sqlx"
	"veza-backend/middleware"
)

func RegisterSearchRoutes(r *mux.Router, db *sqlx.DB) {
	// Recherche globale - publique ou protégée selon votre logique
	r.Handle("/search", handlers.GlobalSearchHandler(db)).Methods("GET")
	
	// Recherche avancée avec filtres - nécessite une authentification
	r.Handle("/search/advanced", middleware.JWTAuthMiddleware(handlers.AdvancedSearchHandler(db))).Methods("GET")
	
	// Auto-complétion
	r.Handle("/autocomplete", handlers.AutocompleteHandler(db)).Methods("GET")
}