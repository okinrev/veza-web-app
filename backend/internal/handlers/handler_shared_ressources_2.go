// file: backend/handlers/shared_ressources.go

package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"log"
	"fmt"
	"mime"
	"path/filepath"
	"strings"
	"time"

	"veza-web-app/db"
	"veza-web-app/models"
	"veza-web-app/middleware"

	"github.com/lib/pq"
	"github.com/gorilla/mux"
)

func UploadSharedResource(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(32 << 20) // 32MB
	if err != nil {
		http.Error(w, "invalid form", http.StatusBadRequest)
		return
	}

	title := r.FormValue("title")
	typeStr := r.FormValue("type") // e.g. sample, preset
	tagsStr := r.FormValue("tags")
	description := r.FormValue("description")
	isPublic := r.FormValue("is_public") != "false" // true par défaut
	tags := []string{}
	if tagsStr != "" {
		tags = strings.Split(tagsStr, ",")
		for i := range tags {
			tags[i] = strings.TrimSpace(tags[i])
		}
	}

	uploaderID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "missing file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	filename := filepath.Base(handler.Filename)
	savePath := filepath.Join("shared_ressources", filename)
	os.MkdirAll("shared_ressources", os.ModePerm)
	out, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	url := "/shared_ressources/" + filename
	_, err = db.DB.Exec(`
		INSERT INTO shared_ressources (title, description, filename, url, type, tags, uploader_id, is_public, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, title, description, filename, url, typeStr, pq.Array(tags), uploaderID, isPublic, time.Now())
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
		"filename": filename,
		"url": url,
	})
}

func ListSharedResources(w http.ResponseWriter, r *http.Request) {
	var resources []models.SharedResource
	err := db.DB.Select(&resources, `
		SELECT sr.*, u.username AS uploader_username
		FROM shared_ressources sr
		JOIN users u ON sr.uploader_id = u.id
		WHERE sr.is_public = true
		ORDER BY sr.uploaded_at DESC
	  `)
	if err != nil {
		log.Printf("❌ Erreur requête SELECT shared_ressources: %v", err) // <== ajoute ce log
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resources)
}

func ServeSharedFile(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	path := filepath.Join("shared_ressources", filename)
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	
	if r.URL.Query().Get("download") == "true" {
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))

		// 🔐 Force le bon type MIME
		ext := filepath.Ext(filename)
		mimeType := mime.TypeByExtension(ext)
		if mimeType != "" {
			w.Header().Set("Content-Type", mimeType)
		} else {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		_, err := db.DB.Exec(`UPDATE shared_ressources SET download_count = download_count + 1 WHERE filename = $1`, filename)
		if err != nil {
			log.Printf("❌ Erreur mise à jour download_count : %v", err)
		}
	}

	http.ServeFile(w, r, path)
}

func SearchSharedResources(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	resType := r.URL.Query().Get("type")
	title := r.URL.Query().Get("title")
	uploader := r.URL.Query().Get("uploader")

	args := []interface{}{}
	i := 1

	// Choix du SELECT + JOIN uniquement si nécessaire
	selectFields := "sr.*"
	joinClause := ""
	if uploader != "" {
		selectFields += ", u.username AS uploader_username"
		joinClause = "JOIN users u ON sr.uploader_id = u.id"
	}

	query := fmt.Sprintf(`
		SELECT %s
		FROM shared_ressources sr
		%s
		WHERE sr.is_public = true
	`, selectFields, joinClause)

	if tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(sr.tags)", i)
		args = append(args, tag)
		i++
	}
	if resType != "" {
		query += fmt.Sprintf(" AND LOWER(sr.type) = LOWER($%d)", i)
		args = append(args, resType)
		i++
	}
	if title != "" {
		query += fmt.Sprintf(" AND LOWER(sr.title) ILIKE LOWER($%d)", i)
		args = append(args, "%"+title+"%")
		i++
	}
	if uploader != "" {
		query += fmt.Sprintf(" AND LOWER(u.username) ILIKE LOWER($%d)", i)
		args = append(args, "%"+uploader+"%")
		i++
	}

	query += " ORDER BY sr.uploaded_at DESC"

	var ressources []models.SharedResource
	err := db.DB.Select(&ressources, query, args...)
	if err != nil {
		log.Printf("❌ Erreur recherche: %v", err)
		http.Error(w, "Erreur base de données", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(ressources)
}


func DeleteSharedResource(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]
	_, err := db.DB.Exec(`DELETE FROM shared_ressources WHERE id = $1`, id)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"deleted"}`))
}

func UpdateSharedResource(w http.ResponseWriter, r *http.Request) {
	id := mux.Vars(r)["id"]

	var req struct {
		Title       *string         `json:"title"`
		Description *string         `json:"description"`
		Type        *string         `json:"type"`
		Tags        []string        `json:"tags"`
		IsPublic    *bool           `json:"is_public"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	query := `UPDATE shared_ressources SET `
	args := []interface{}{}
	i := 1

	if req.Title != nil {
		query += fmt.Sprintf("title = $%d, ", i)
		args = append(args, *req.Title)
		i++
	}
	if req.Description != nil {
		query += fmt.Sprintf("description = $%d, ", i)
		args = append(args, *req.Description)
		i++
	}
	if req.Type != nil {
		query += fmt.Sprintf("type = $%d, ", i)
		args = append(args, *req.Type)
		i++
	}
	if req.Tags != nil {
		query += fmt.Sprintf("tags = $%d, ", i)
		args = append(args, pq.Array(req.Tags))
		i++
	}
	if req.IsPublic != nil {
		query += fmt.Sprintf("is_public = $%d, ", i)
		args = append(args, *req.IsPublic)
		i++
	}

	if len(args) == 0 {
		http.Error(w, "no fields to update", http.StatusBadRequest)
		return
	}

	query = strings.TrimSuffix(query, ", ")
	query += fmt.Sprintf(" WHERE id = $%d", i)
	args = append(args, id)

	_, err := db.DB.Exec(query, args...)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"updated"}`))
}
