// file: routes/exchange.go

 package routes

// import (
// 	"github.com/gorilla/mux"
// 	"veza-web-app/handlers"
// 	"veza-web-app/middleware"
// )

// func RegisterExchangeRoutes(r *mux.Router) {
// 	protected := r.PathPrefix("/exchange").Subrouter()
// 	protected.Use(middleware.JWTAuthMiddleware)

// 	protected.HandleFunc("", handlers.CreateExchangeOffer).Methods("POST")
// 	protected.HandleFunc("", handlers.ListExchangeOffers).Methods("GET")
// 	protected.HandleFunc("/{id}/accept", handlers.AcceptExchangeOffer).Methods("POST")
// 	protected.HandleFunc("/{id}/cancel", handlers.CancelExchangeOffer).Methods("POST")
// }
