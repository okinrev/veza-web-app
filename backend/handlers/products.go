//file: backend/handlers/products.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"

	"veza-backend/db"
	"veza-backend/middleware"
	"veza-backend/models"
)

func ListUserProducts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var products []models.Product
	err := db.DB.Select(&products, `
		SELECT id, name, version, purchase_date, warranty_expires
		FROM products WHERE user_id = $1
		ORDER BY purchase_date DESC
	`, userID)

	if err != nil {
		http.Error(w, "Erreur lors de la récupération des produits", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

func GetProductDetails(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var product models.Product
	err = db.DB.Get(&product, `
		SELECT id, user_id, name, version, purchase_date, warranty_expires
		FROM products WHERE id = $1
	`, productID)

	if err != nil {
		http.Error(w, "Produit introuvable", http.StatusNotFound)
		return
	}

	if product.UserID != userID {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(product)
}

func CreateProduct(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	p.UserID = userID
	err := db.DB.QueryRowx(`
		INSERT INTO products (user_id, name, version, purchase_date, warranty_expires)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, p.UserID, p.Name, p.Version, p.PurchaseDate, p.WarrantyExpires).Scan(&p.ID)

	if err != nil {
		http.Error(w, "Erreur lors de la création", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func UpdateProduct(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	var ownerID int
	err = db.DB.Get(&ownerID, "SELECT user_id FROM products WHERE id = $1", id)
	if err != nil || ownerID != userID {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	_, err = db.DB.Exec(`
		UPDATE products SET name = $1, version = $2, purchase_date = $3, warranty_expires = $4
		WHERE id = $5
	`, p.Name, p.Version, p.PurchaseDate, p.WarrantyExpires, id)

	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func DeleteProduct(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var ownerID int
	err = db.DB.Get(&ownerID, "SELECT user_id FROM products WHERE id = $1", id)
	if err != nil || ownerID != userID {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	_, err = db.DB.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
