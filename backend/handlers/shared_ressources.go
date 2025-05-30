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

	"veza-backend/db"
	"veza-backend/models"
	"veza-backend/middleware"

	"github.com/lib/pq"
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
		INSERT INTO shared_ressources (title, filename, url, type, tags, uploader_id, is_public, uploaded_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`, title, filename, url, typeStr, pq.Array(tags), uploaderID, true, time.Now())
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
		SELECT * FROM shared_ressources
		WHERE is_public = true
		ORDER BY uploaded_at DESC`)
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
	}

	http.ServeFile(w, r, path)
}

func SearchSharedResources(w http.ResponseWriter, r *http.Request) {
	tag := r.URL.Query().Get("tag")
	resType := r.URL.Query().Get("type")
	title := r.URL.Query().Get("title")

	query := `SELECT * FROM shared_ressources WHERE is_public = true`
	args := []interface{}{}
	i := 1

	if tag != "" {
		query += fmt.Sprintf(" AND $%d = ANY(tags)", i)
		args = append(args, tag)
		i++
	}
	if resType != "" {
		query += fmt.Sprintf(" AND LOWER(type) = LOWER($%d)", i)
		args = append(args, resType)
		i++
	}
	if title != "" {
		query += fmt.Sprintf(" AND LOWER(title) ILIKE LOWER($%d)", i)
		args = append(args, "%"+title+"%")
		i++
	}

	query += " ORDER BY uploaded_at DESC"

	var ressources []models.SharedResource
	err := db.DB.Select(&ressources, query, args...)
	if err != nil {
		log.Printf("❌ Erreur recherche: %v", err)
		http.Error(w, "Erreur base de données", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(ressources)
}
