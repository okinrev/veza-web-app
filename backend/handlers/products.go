//file: backend/handlers/admin_products.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"veza-backend/db"
	"veza-backend/middleware"
	"veza-backend/models"
)

// ListAdminProducts - Liste tous les produits avec informations détaillées pour l'admin
func ListAdminProducts(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	
	query := `
		SELECT 
			p.id, p.name, p.category_id, p.brand, p.model, p.description, 
			p.price, p.warranty_months, p.warranty_conditions, p.manufacturer_website,
			p.specifications, p.status, p.created_at, p.updated_at,
			COALESCE(c.name, '') as category_name,
			COALESCE(doc_count.count, 0) as documentation_count,
			COALESCE(user_count.count, 0) as user_count
		FROM products p
		LEFT JOIN categories c ON p.category_id = c.id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as count 
			FROM product_documents 
			GROUP BY product_id
		) doc_count ON p.id = doc_count.product_id
		LEFT JOIN (
			SELECT product_id, COUNT(*) as count 
			FROM user_products 
			GROUP BY product_id
		) user_count ON p.id = user_count.product_id
		ORDER BY p.updated_at DESC
	`
	
	err := db.DB.Select(&products, query)
	if err != nil {
		http.Error(w, "Erreur lors de la récupération des produits", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// CreateAdminProduct - Créer un nouveau produit (admin)
func CreateAdminProduct(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	
	// TODO: Vérifier si l'utilisateur est admin
	_ = userID // Pour éviter l'erreur de variable non utilisée
	
	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	// Validation des données
	if req.Name == "" {
		http.Error(w, "Le nom du produit est obligatoire", http.StatusBadRequest)
		return
	}

	if req.Status == "" {
		req.Status = "active"
	}

	var product models.Product
	err := db.DB.QueryRowx(`
		INSERT INTO products (
			name, category_id, brand, model, description, price, 
			warranty_months, warranty_conditions, manufacturer_website, 
			specifications, status, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13) 
		RETURNING id, name, category_id, brand, model, description, price, 
				  warranty_months, warranty_conditions, manufacturer_website, 
				  specifications, status, created_at, updated_at
	`, req.Name, req.CategoryID, req.Brand, req.Model, req.Description, req.Price,
		req.WarrantyMonths, req.WarrantyConditions, req.ManufacturerWebsite,
		req.Specifications, req.Status, time.Now(), time.Now()).StructScan(&product)

	if err != nil {
		http.Error(w, "Erreur lors de la création: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// UpdateAdminProduct - Mettre à jour un produit (admin)
func UpdateAdminProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var req models.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(`
		UPDATE products SET 
			name = $1, category_id = $2, brand = $3, model = $4, description = $5,
			price = $6, warranty_months = $7, warranty_conditions = $8, 
			manufacturer_website = $9, specifications = $10, status = $11, updated_at = $12
		WHERE id = $13
	`, req.Name, req.CategoryID, req.Brand, req.Model, req.Description, req.Price,
		req.WarrantyMonths, req.WarrantyConditions, req.ManufacturerWebsite,
		req.Specifications, req.Status, time.Now(), id)

	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteAdminProduct - Supprimer un produit (admin)
func DeleteAdminProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	// Vérifier si le produit est utilisé
	var count int
	err = db.DB.Get(&count, `
		SELECT COUNT(*) FROM user_products WHERE product_id = $1
	`, id)

	if err == nil && count > 0 {
		http.Error(w, "Impossible de supprimer : produit utilisé par des utilisateurs", http.StatusConflict)
		return
	}

	_, err = db.DB.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

//file: backend/handlers/admin_categories.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"veza-backend/db"
	"veza-backend/models"
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