//file: backend/routes/offer.go

package routes

import (
    "github.com/gorilla/mux"
    "veza-backend/handlers"
)

func RegisterOfferRoutes(router *mux.Router) {
    // Offers
    router.HandleFunc("/listings/{id:[0-9]+}/offer", handlers.CreateOffer).Methods("POST")
    router.HandleFunc("/offers/{id:[0-9]+}/accept", handlers.AcceptOffer).Methods("POST")
}
