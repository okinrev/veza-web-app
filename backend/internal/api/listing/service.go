//file: backend/routes/listing.go

package routes

import (
    "github.com/gorilla/mux"
    "backend/handlers"
)

func RegisterListingRoutes(router *mux.Router) {
    // CRUD Listings
    router.HandleFunc("/listings", handlers.CreateListing).Methods("POST")
    router.HandleFunc("/listings", handlers.GetAllListings).Methods("GET")
    router.HandleFunc("/listings/{id:[0-9]+}", handlers.GetListingByID).Methods("GET")
    router.HandleFunc("/listings/{id:[0-9]+}", handlers.DeleteListing).Methods("DELETE")
}
