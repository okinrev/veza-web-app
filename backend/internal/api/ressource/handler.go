//file: backend/handlers/ressource.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"fmt"
	"path/filepath"
	"os"

	"github.com/gorilla/mux"
	"backend/db"
	"backend/middleware"
	"backend/models"
)

func ListInternalDocs(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	productID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	// Vérifie l’accès à ce produit
	var ownerID int
	err = db.DB.Get(&ownerID, "SELECT user_id FROM products WHERE id = $1", productID)
	if err != nil || ownerID != userID {
		http.Error(w, "Accès non autorisé", http.StatusUnauthorized)
		return
	}

	var docs []models.InternalRessource
	err = db.DB.Select(&docs, "SELECT * FROM internal_documents WHERE product_id = $1 ORDER BY uploaded_at DESC", productID)
	if err != nil {
		http.Error(w, "Erreur récupération documents", http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(docs)
}

func ServeInternalDocByID(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var doc models.InternalRessource
	err = db.DB.Get(&doc, "SELECT * FROM internal_documents WHERE id = $1", docID)
	if err != nil {
		http.Error(w, "Document introuvable", http.StatusNotFound)
		return
	}

	// Vérifie si le fichier existe réellement
	safeName := filepath.Base(doc.URL) // protection path traversal
	fullPath := filepath.Join("internal_docs", safeName)

	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		http.Error(w, "Fichier introuvable", http.StatusNotFound)
		return
	}

	// Force le téléchargement
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", doc.Filename))
	w.Header().Set("Content-Type", "application/pdf") // ou autre selon le type
	http.ServeFile(w, r, fullPath)
}

func GetInternalRessource(w http.ResponseWriter, r *http.Request) {
	docID, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var doc models.InternalRessource
	err = db.DB.Get(&doc, "SELECT * FROM internal_documents WHERE id = $1", docID)
	if err != nil {
		http.Error(w, "Document introuvable", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(doc)
}
