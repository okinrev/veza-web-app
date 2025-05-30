//file: backend/handlers/file.go

package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"encoding/json"

	"github.com/gorilla/mux"
	"veza-backend/db"
	"veza-backend/middleware"
	"veza-backend/models"
)

func UploadFileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	productID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	// Vérifie que le produit appartient à l'utilisateur
	var ownerID int
	err = db.DB.Get(&ownerID, "SELECT user_id FROM products WHERE id = $1", productID)
	if err != nil || ownerID != userID {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	// Parse le fichier reçu (multipart form)
	err = r.ParseMultipartForm(10 << 20) // 10 MB
	if err != nil {
		http.Error(w, "Fichier trop volumineux", http.StatusBadRequest)
		return
	}

	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Fichier manquant", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Récupère le type
	fileType := r.FormValue("type")
	if fileType == "" {
		fileType = "manual" // fallback
	}

	// Nom et chemin de destination
	safeName := fmt.Sprintf("%d_%d_%s", productID, time.Now().Unix(), strings.ReplaceAll(handler.Filename, " ", "_"))
	savePath := filepath.Join("uploads", safeName)

	dst, err := os.Create(savePath)
	if err != nil {
		http.Error(w, "Erreur d'écriture fichier", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		http.Error(w, "Erreur lors de l'enregistrement", http.StatusInternalServerError)
		return
	}

	// Enregistrement en base
	_, err = db.DB.Exec(`
		INSERT INTO files (product_id, filename, url, type)
		VALUES ($1, $2, $3, $4)
	`, productID, handler.Filename, "/files/"+safeName, fileType)
	if err != nil {
		http.Error(w, "Erreur DB", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"message":"Fichier uploadé avec succès"}`))
}

func DownloadFileHandler(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	fileID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var file models.File
	err = db.DB.Get(&file, "SELECT * FROM files WHERE id = $1", fileID)
	if err != nil {
		http.Error(w, "Fichier introuvable", http.StatusNotFound)
		return
	}

	// Vérifie l'accès au produit
	var ownerID int
	err = db.DB.Get(&ownerID, `
		SELECT user_id FROM products WHERE id = $1
	`, file.ProductID)

	if err != nil || ownerID != userID {
		http.Error(w, "Accès non autorisé", http.StatusUnauthorized)
		return
	}

	// Récupère le vrai chemin
	filename := filepath.Base(file.URL) // sécurité : empêche path traversal
	fullPath := filepath.Join("uploads", filename)

	// Envoie le header pour forcer le téléchargement
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", file.Filename))
	w.Header().Set("Content-Type", "application/octet-stream")

	// Envoie le fichier
	http.ServeFile(w, r, fullPath)
}

func ListProductFiles(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	productID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	// Vérifie l'accès au produit
	var ownerID int
	err = db.DB.Get(&ownerID, "SELECT user_id FROM products WHERE id = $1", productID)
	if err != nil || ownerID != userID {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	// Récupère les fichiers associés
	var files []models.File
	err = db.DB.Select(&files, "SELECT * FROM files WHERE product_id = $1 ORDER BY uploaded_at DESC", productID)
	if err != nil {
		http.Error(w, "Erreur récupération fichiers", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(files)
}
