//file: backend/handlers/admin_categories.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"veza-web-app/db"
	"veza-web-app/models"
)

// ListAdminCategories - Liste toutes les catégories avec le nombre de produits
func ListAdminCategories(w http.ResponseWriter, r *http.Request) {
	var categories []models.Category
	
	query := `
		SELECT 
			c.id, c.name, c.description, c.created_at, c.updated_at,
			COALESCE(p.count, 0) as product_count
		FROM categories c
		LEFT JOIN (
			SELECT category_id, COUNT(*) as count 
			FROM products 
			WHERE category_id IS NOT NULL
			GROUP BY category_id
		) p ON c.id = p.category_id
		ORDER BY c.name ASC
	`
	
	err := db.DB.Select(&categories, query)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des catégories", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

// CreateAdminCategory - Créer une nouvelle catégorie
func CreateAdminCategory(w http.ResponseWriter, r *http.Request) {
	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	if category.Name == "" {
		http.Error(w, "Le nom de la catégorie est obligatoire", http.StatusBadRequest)
		return
	}

	err := db.DB.QueryRowx(`
		INSERT INTO categories (name, description, created_at, updated_at) 
		VALUES ($1, $2, $3, $4) 
		RETURNING id, name, description, created_at, updated_at
	`, category.Name, category.Description, time.Now(), time.Now()).StructScan(&category)

	if err != nil {
		http.Error(w, "Erreur lors de la création: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(category)
}

// UpdateAdminCategory - Mettre à jour une catégorie
func UpdateAdminCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var category models.Category
	if err := json.NewDecoder(r.Body).Decode(&category); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(`
		UPDATE categories SET name = $1, description = $2, updated_at = $3 WHERE id = $4
	`, category.Name, category.Description, time.Now(), id)

	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteAdminCategory - Supprimer une catégorie
func DeleteAdminCategory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	// Vérifier si la catégorie est utilisée
	var count int
	err = db.DB.Get(&count, `SELECT COUNT(*) FROM products WHERE category_id = $1`, id)
	if err == nil && count > 0 {
		http.Error(w, "Impossible de supprimer : catégorie utilisée par des produits", http.StatusConflict)
		return
	}

	_, err = db.DB.Exec("DELETE FROM categories WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}