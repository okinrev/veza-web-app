//file: backend/handlers/room.go

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"backend/models" // ← Ajoute cet import pour accéder aux structs Room et Message
)

// 🔍 GET /chat/rooms — Liste des salons publics
func GetPublicRoomsHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var rooms []models.Room

		err := db.Select(&rooms, `
            SELECT id, name, is_private, created_at 
            FROM rooms 
            WHERE is_private = false 
            ORDER BY created_at DESC
        `)
		if err != nil {
			http.Error(w, "Erreur DB", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(rooms)
	}
}

// ➕ POST /chat/rooms — Création d’un salon
func CreateRoomHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Name      string `json:"name"`
			IsPrivate bool   `json:"is_private"`
		}

		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Données invalides", http.StatusBadRequest)
			return
		}

		_, err := db.Exec(`
            INSERT INTO rooms (name, is_private) VALUES ($1, $2)
        `, input.Name, input.IsPrivate)
		if err != nil {
			http.Error(w, "Erreur DB", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"Salon créé"}`))
	}
}

// 📜 GET /chat/rooms/{room}/messages — Historique d’un salon
func GetRoomMessagesHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		room := mux.Vars(r)["room"]

		var messages []models.Message

		err := db.Select(&messages, `
            SELECT id, from_user, content, timestamp 
            FROM messages 
            WHERE room = $1 
            ORDER BY timestamp DESC 
            LIMIT 50
        `, room)
		if err != nil {
			http.Error(w, "Erreur DB", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(messages)
	}
}
