//file: backend/routes/message.go

package routes

import (
    "github.com/gorilla/mux"
    "veza-backend/handlers"
    "github.com/jmoiron/sqlx"
    "veza-backend/middleware"
)

func RegisterMessageRoutes(r *mux.Router, db *sqlx.DB) {
	r.Handle("/chat/dm/{user_id}", middleware.JWTAuthMiddleware(handlers.GetDmHandler(db))).Methods("GET")
}
