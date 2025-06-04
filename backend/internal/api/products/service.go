//file: backend/routes/products.go

package routes

import (
	"github.com/gorilla/mux"
	"veza-web-app/handlers"
	"veza-web-app/middleware"
)

func RegisterAdminRoutes(r *mux.Router) {
	// Toutes les routes admin nécessitent une authentification
	admin := r.PathPrefix("/admin").Subrouter()
	admin.Use(middleware.JWTAuthMiddleware)
	
	// TODO: Ajouter un middleware pour vérifier les droits admin
	// admin.Use(middleware.AdminAuthMiddleware)
	
	// Routes pour les produits
	admin.HandleFunc("/products", handlers.ListAdminProducts).Methods("GET")
	admin.HandleFunc("/products", handlers.CreateAdminProduct).Methods("POST")
	admin.HandleFunc("/products/{id:[0-9]+}", handlers.UpdateAdminProduct).Methods("PUT")
	admin.HandleFunc("/products/{id:[0-9]+}", handlers.DeleteAdminProduct).Methods("DELETE")
	
	// Routes pour les catégories
	admin.HandleFunc("/categories", handlers.ListAdminCategories).Methods("GET")
	admin.HandleFunc("/categories", handlers.CreateAdminCategory).Methods("POST")
	admin.HandleFunc("/categories/{id:[0-9]+}", handlers.UpdateAdminCategory).Methods("PUT")
	admin.HandleFunc("/categories/{id:[0-9]+}", handlers.DeleteAdminCategory).Methods("DELETE")
	
	// Routes pour les documents (à implémenter)
	// admin.HandleFunc("/products/{id:[0-9]+}/documents", handlers.ListProductDocuments).Methods("GET")
	// admin.HandleFunc("/products/{id:[0-9]+}/documents", handlers.UploadProductDocument).Methods("POST")
	// admin.HandleFunc("/documents/{id:[0-9]+}", handlers.DeleteProductDocument).Methods("DELETE")
}