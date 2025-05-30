// file: backend/handlers/tag.go
package handlers

import (
	"encoding/json"
	"net/http"
	"strings"
	"log"

	"veza-backend/db"
)

func GetAllTags(w http.ResponseWriter, r *http.Request) {
	rows, err := db.DB.Query("SELECT name FROM tags ORDER BY name")
	if err != nil {
		log.Println("Erreur DB:", err)
		http.Error(w, "Erreur lors de la récupération des tags", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tags = append(tags, name)
		}
	}
	json.NewEncoder(w).Encode(tags)
}

func SearchTags(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	if q == "" {
		log.Println("Erreur DB:", q)

		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	rows, err := db.DB.Query("SELECT name FROM tags WHERE name ILIKE $1 ORDER BY name LIMIT 20", q+"%")
	if err != nil {
		http.Error(w, "Erreur DB", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err == nil {
			tags = append(tags, name)
		}
	}
	json.NewEncoder(w).Encode(tags)
}