// file: backend/handlers/shared_ressources.go

package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
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
	savePath := filepath.Join("shared", filename)
	os.MkdirAll("shared", os.ModePerm)
	out, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "cannot save file", http.StatusInternalServerError)
		return
	}
	defer out.Close()
	io.Copy(out, file)

	url := "/shared/" + filename
	_, err = db.DB.Exec(`
		INSERT INTO shared_resources (title, filename, url, type, tags, uploader_id, is_public, uploaded_at)
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
		SELECT * FROM shared_resources
		WHERE is_public = true
		ORDER BY uploaded_at DESC`)
	if err != nil {
		http.Error(w, "db error", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(resources)
}

func ServeSharedFile(w http.ResponseWriter, r *http.Request) {
	filename := filepath.Base(r.URL.Path)
	path := filepath.Join("shared", filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}
	http.ServeFile(w, r, path)
}
