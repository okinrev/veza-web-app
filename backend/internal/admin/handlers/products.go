package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	
	"github.com/gorilla/mux"
	"veza-backend/internal/utils/response"
	"veza-backend/db"
	"veza-backend/models"
)

type ProductHandler struct{}

func NewProductHandler() *ProductHandler {
	return &ProductHandler{}
}

// Adaptation temporaire des handlers existants avec les nouvelles réponses
func (h *ProductHandler) GetProducts(w http.ResponseWriter, r *http.Request) {
	var products []models.Product
	err := db.DB.Select(&products, `
		SELECT p.id, p.name, 
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
		ORDER BY p.id DESC
	`)
	
	if err != nil {
		response.ErrorJSON(w, "Erreur lors de la récupération des produits", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(w, products, "Produits récupérés avec succès")
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.ErrorJSON(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	err := db.DB.QueryRowx(`
		INSERT INTO products (name) VALUES ($1) RETURNING id
	`, p.Name).Scan(&p.ID)

	if err != nil {
		response.ErrorJSON(w, "Erreur lors de la création", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(w, p, "Produit créé avec succès")
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.ErrorJSON(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		response.ErrorJSON(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec(`UPDATE products SET name = $1 WHERE id = $2`, p.Name, id)
	if err != nil {
		response.ErrorJSON(w, "Erreur lors de la mise à jour", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(w, nil, "Produit mis à jour avec succès")
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		response.ErrorJSON(w, "ID invalide", http.StatusBadRequest)
		return
	}

	_, err = db.DB.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		response.ErrorJSON(w, "Erreur lors de la suppression", http.StatusInternalServerError)
		return
	}

	response.SuccessJSON(w, nil, "Produit supprimé avec succès")
}
