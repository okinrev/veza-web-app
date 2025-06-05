//file: backend/routes/room.go

package routes

import (
    "github.com/gorilla/mux"
    "veza-backend/handlers"
    "github.com/jmoiron/sqlx"
    "veza-backend/middleware"
)

func RegisterRoomRoutes(r *mux.Router, db *sqlx.DB) {
	r.Handle("/chat/rooms", middleware.JWTAuthMiddleware(handlers.GetPublicRoomsHandler(db))).Methods("GET")
	r.Handle("/chat/rooms", middleware.JWTAuthMiddleware(handlers.CreateRoomHandler(db))).Methods("POST")
	r.Handle("/chat/rooms/{room}/messages", middleware.JWTAuthMiddleware(handlers.GetRoomMessagesHandler(db))).Methods("GET")
}
