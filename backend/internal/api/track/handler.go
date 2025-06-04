//file: backend/handlers/tracks.go

package handlers

import (
	"encoding/json"
	"net/http"
	"log"
	"io"
	"os"
    "strings"
	"path/filepath"
    "github.com/lib/pq"
    "github.com/gorilla/mux"

	"backend/db"
	"backend/models"
    "backend/middleware"
    "backend/utils"
)

func ListTracks(w http.ResponseWriter, r *http.Request) {
	var tracks []models.Track
	err := db.DB.Select(&tracks, `SELECT id, title, artist, filename, created_at, duration_seconds, tags, is_public, uploader_id FROM tracks`)
	log.Printf("Tracks récupérées : %v", tracks)
	if err != nil {
        log.Printf("DB error: %v", err) // ← ajoute ceci
		http.Error(w, "Erreur récupération des pistes", http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(tracks)
}

func AddTrackWithUpload(w http.ResponseWriter, r *http.Request) {
    err := r.ParseMultipartForm(32 << 20) // 32MB max
    if err != nil {
        http.Error(w, "invalid multipart form", http.StatusBadRequest)
        return
    }

    userID, ok := r.Context().Value(middleware.UserIDKey).(int)
    if !ok {
        http.Error(w, "Utilisateur non authentifié", http.StatusUnauthorized)
        return
    }


    title := r.FormValue("title")
    artist := r.FormValue("artist")
    tagsStr := r.FormValue("tags") // ex: "hiphop,lofi,night"
    var tags []string
    if tagsStr != "" {
        tags = strings.Split(tagsStr, ",")
        for i := range tags {
            tags[i] = strings.TrimSpace(tags[i]) // nettoie les espaces
        }
    } else {
        tags = []string{} // tableau vide propre
    }

    log.Printf("tags added → %v", tags)

    file, handler, err := r.FormFile("audio")
    if err != nil {
        http.Error(w, "audio file missing", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Enregistre dans audio/
    filename := filepath.Base(handler.Filename)
	log.Printf("filename → %s", filename)
    savePath := filepath.Join("audio", filename)
	log.Printf("save path → %s", savePath)

	err = os.MkdirAll("audio", os.ModePerm)
	if err != nil {
		http.Error(w, "failed to create audio directory", http.StatusInternalServerError)
		return
	}


    out, err := os.Create(savePath)
	log.Printf("err → %s", err)

    if err != nil {
        http.Error(w, "failed to save file", http.StatusInternalServerError)
        return
    }
    defer out.Close()

    _, err = io.Copy(out, file)
    if err != nil {
        http.Error(w, "failed to write file", http.StatusInternalServerError)
        return
    }

    log.Printf("inserting track: %s by %s → %s", title, artist, filename)

    // après io.Copy()
    _, err = db.DB.Exec(`
        INSERT INTO tracks (title, artist, filename, duration_seconds, tags, is_public, uploader_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
    `, title, artist, filename, 0, pq.Array(tags), true, userID)
    if err != nil {
        log.Printf("DB insert error → %s", err)
        http.Error(w, "DB insert failed", http.StatusInternalServerError)
        return
    }

    // Simule une insertion en base
    log.Printf("🎵 Track added: %s by %s → %s", title, artist, savePath)

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "status": "ok",
        "title": title,
        "artist": artist,
        "path": savePath,
    })
}

// GET /stream/{filename}
func StreamAudio(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    filename := vars["filename"]

    if filename == "" {
        http.NotFound(w, r)
        return
    }

    path := filepath.Join("audio", filename)

    if _, err := os.Stat(path); os.IsNotExist(err) {
        http.NotFound(w, r)
        return
    }

    http.ServeFile(w, r, path)
}


func StreamAudioSigned(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	filename := vars["filename"]
	expires := r.URL.Query().Get("expires")
	sig := r.URL.Query().Get("sig")

	if !utils.ValidateSignature(filename, expires, sig) {
		http.Error(w, "Lien invalide ou expiré", http.StatusUnauthorized)
		return
	}

	path := filepath.Join("audio", filename)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.NotFound(w, r)
		return
	}

	http.ServeFile(w, r, path)
}