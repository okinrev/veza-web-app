//file: backend/handlers/message.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// 📩 GET /chat/dm/{user_id} — Historique d’un DM entre deux utilisateurs
// GET /chat/dm/{user_id} — Historique d’un DM entre deux utilisateurs
func GetDmHandler(db *sqlx.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		currentUserID, ok := r.Context().Value("user_id").(int)
		if !ok || currentUserID == 0 {
			http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
			return
		}

		otherUserID, err := strconv.Atoi(mux.Vars(r)["user_id"])
		if err != nil {
			http.Error(w, "ID utilisateur invalide", http.StatusBadRequest)
			return
		}

		// Vérifie que le destinataire existe
		var count int
		err = db.Get(&count, `SELECT COUNT(*) FROM users WHERE id = $1`, otherUserID)
		if err != nil || count == 0 {
			http.Error(w, "Utilisateur introuvable", http.StatusNotFound)
			return
		}

		// Nouveau type pour réponse enrichie
		type DMWithUsername struct {
			ID        int    `json:"id"`
			FromUser  int    `db:"from_user" json:"fromUser"`
			To        int    `db:"to_user" json:"to"`
			Content   string `json:"content"`
			Timestamp string `json:"timestamp"`
			Username  string `json:"username"`
		}

		var messages []DMWithUsername
		err = db.Select(&messages, `
			SELECT m.id, m.from_user, m.to_user, m.content, m.timestamp, u.username
			FROM messages m
			JOIN users u ON u.id = m.from_user
			WHERE (m.from_user = $1 AND m.to_user = $2)
			   OR (m.from_user = $2 AND m.to_user = $1)
			ORDER BY m.timestamp ASC
			LIMIT 50
		`, currentUserID, otherUserID)

		if err != nil {
			http.Error(w, "Erreur DB", http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(messages)
	}
}
