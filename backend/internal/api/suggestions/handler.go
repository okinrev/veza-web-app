// file: backend/handlers/suggestions.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"backend/db"
)

func GetSuggestions(w http.ResponseWriter, r *http.Request) {
	tag := strings.TrimSpace(r.URL.Query().Get("tag"))
	title := strings.TrimSpace(r.URL.Query().Get("title"))
	uploader := strings.TrimSpace(r.URL.Query().Get("uploader"))
	doc := strings.TrimSpace(r.URL.Query().Get("doc"))
	user := strings.TrimSpace(r.URL.Query().Get("user"))
	track := strings.TrimSpace(r.URL.Query().Get("track"))

	var query string
	var param string
	var col string
	var table string

	switch {
	case tag != "":
		table = "tags"
		col = "name"
		param = tag + "%"
	case title != "":
		table = "shared_ressources"
		col = "title"
		param = title + "%"
	case uploader != "":
		table = "shared_ressources"
		col = "uploader_username"
		param = uploader + "%"
	case doc != "":
		table = "internal_ressources"
		col = "title"
		param = doc + "%"
	case user != "":
		table = "users"
		col = "username"
		param = user + "%"
	case track != "":
		table = "tracks"
		col = "title"
		param = track + "%"
	default:
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`[]`))
		return
	}

	query = `SELECT DISTINCT ` + col + ` FROM ` + table + ` WHERE ` + col + ` ILIKE $1 ORDER BY ` + col + ` LIMIT 15`

	rows, err := db.DB.Query(query, param)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]string{})
		return
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var val string
		if err := rows.Scan(&val); err == nil {
			results = append(results, val)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}
