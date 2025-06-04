//file: backend/handlers/user_products.go

package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"github.com/gorilla/mux"
	"backend/db"
	"backend/middleware"
	"backend/models"
)

// ListUserProducts - Liste les produits possédés par l'utilisateur connecté
func ListUserProducts(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var products []models.UserProductWithName
	err := db.DB.Select(&products, `
		SELECT up.id, up.user_id, up.product_id, p.name as product_name, 
		       up.version, up.purchase_date, up.warranty_expires
		FROM user_products up
		JOIN products p ON up.product_id = p.id
		WHERE up.user_id = $1
		ORDER BY up.purchase_date DESC
	`, userID)

	if err != nil {
		http.Error(w, "Erreur lors de la récupération des produits", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(products)
}

// GetUserProductDetails - Détails d'un produit utilisateur spécifique
func GetUserProductDetails(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	productID, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var product models.UserProductWithName
	err = db.DB.Get(&product, `
		SELECT up.id, up.user_id, up.product_id, p.name as product_name,
		       up.version, up.purchase_date, up.warranty_expires
		FROM user_products up
		JOIN products p ON up.product_id = p.id
		WHERE up.id = $1
	`, productID)

	if err != nil {
		http.Error(w, "Produit introuvable", http.StatusNotFound)
		return
	}

	if product.UserID != userID {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(product)
}

// CreateUserProduct - Ajouter un produit à la collection de l'utilisateur
func CreateUserProduct(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var p models.UserProduct
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	// Vérifier que le product_id existe dans la table products
	var exists int
	err := db.DB.Get(&exists, "SELECT COUNT(*) FROM products WHERE id = $1", p.ProductID)
	if err != nil || exists == 0 {
		http.Error(w, "Produit inexistant dans le catalogue", http.StatusBadRequest)
		return
	}

	p.UserID = userID
	err = db.DB.QueryRowx(`
		INSERT INTO user_products (user_id, product_id, version, purchase_date, warranty_expires)
		VALUES ($1, $2, $3, $4, $5) RETURNING id
	`, p.UserID, p.ProductID, p.Version, p.PurchaseDate, p.WarrantyExpires).Scan(&p.ID)

	if err != nil {
		http.Error(w, "Erreur lors de la création", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// UpdateUserProduct - Mettre à jour un produit utilisateur
func UpdateUserProduct(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	var p models.UserProduct
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, "Requête invalide", http.StatusBadRequest)
		return
	}

	// Vérifier que l'utilisateur est propriétaire
	var ownerID int
	err = db.DB.Get(&ownerID, "SELECT user_id FROM user_products WHERE id = $1", id)
	if err != nil || ownerID != userID {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	// Vérifier que le nouveau product_id existe si modifié
	if p.ProductID > 0 {
		var exists int
		err = db.DB.Get(&exists, "SELECT COUNT(*) FROM products WHERE id = $1", p.ProductID)
		if err != nil || exists == 0 {
			http.Error(w, "Produit inexistant dans le catalogue", http.StatusBadRequest)
			return
		}
	}

	_, err = db.DB.Exec(`
		UPDATE user_products SET product_id = $1, version = $2, purchase_date = $3, warranty_expires = $4
		WHERE id = $5
	`, p.ProductID, p.Version, p.PurchaseDate, p.WarrantyExpires, id)

	if err != nil {
		http.Error(w, "Erreur lors de la mise à jour", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// DeleteUserProduct - Supprimer un produit de la collection utilisateur
func DeleteUserProduct(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		http.Error(w, "ID invalide", http.StatusBadRequest)
		return
	}

	// Vérifier que l'utilisateur est propriétaire
	var ownerID int
	err = db.DB.Get(&ownerID, "SELECT user_id FROM user_products WHERE id = $1", id)
	if err != nil || ownerID != userID {
		http.Error(w, "Non autorisé", http.StatusUnauthorized)
		return
	}

	_, err = db.DB.Exec("DELETE FROM user_products WHERE id = $1", id)
	if err != nil {
		http.Error(w, "Erreur lors de la suppression", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}