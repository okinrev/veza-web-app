//file: backend/routes/message.go

package routes

import (
    "github.com/gorilla/mux"
    "veza-web-app/handlers"
    "github.com/jmoiron/sqlx"
    "veza-web-app/middleware"
)

func RegisterMessageRoutes(r *mux.Router, db *sqlx.DB) {
	r.Handle("/chat/dm/{user_id}", middleware.JWTAuthMiddleware(handlers.GetDmHandler(db))).Methods("GET")
}
