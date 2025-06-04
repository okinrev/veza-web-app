//file: backend/models/user_products.go

package models

import "time"

type UserProduct struct {
	ID              int       `db:"id" json:"id"`
	UserID          int       `db:"user_id" json:"user_id"`
	ProductID       int       `db:"product_id" json:"product_id"` // FK vers products
	Version         string    `db:"version" json:"version"`
	PurchaseDate    time.Time `db:"purchase_date" json:"purchase_date"`
	WarrantyExpires time.Time `db:"warranty_expires" json:"warranty_expires"`
}

// UserProductWithName - Structure pour les requêtes avec JOIN
type UserProductWithName struct {
	ID              int       `db:"id" json:"id"`
	UserID          int       `db:"user_id" json:"user_id"`
	ProductID       int       `db:"product_id" json:"product_id"`
	ProductName     string    `db:"product_name" json:"product_name"` // Nom du produit depuis la table products
	Version         string    `db:"version" json:"version"`
	PurchaseDate    time.Time `db:"purchase_date" json:"purchase_date"`
	WarrantyExpires time.Time `db:"warranty_expires" json:"warranty_expires"`
}