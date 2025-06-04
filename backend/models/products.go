//file: backend/models/products.go

package models

import "time"

// Product - Modèle étendu pour l'administration
type Product struct {
	ID                   int       `db:"id" json:"id"`
	Name                 string    `db:"name" json:"name"`
	CategoryID           *int      `db:"category_id" json:"category_id"`
	Brand                string    `db:"brand" json:"brand"`
	Model                string    `db:"model" json:"model"`
	Description          string    `db:"description" json:"description"`
	Price                *float64  `db:"price" json:"price"`
	WarrantyMonths       int       `db:"warranty_months" json:"warranty_months"`
	WarrantyConditions   string    `db:"warranty_conditions" json:"warranty_conditions"`
	ManufacturerWebsite  string    `db:"manufacturer_website" json:"manufacturer_website"`
	Specifications       string    `db:"specifications" json:"specifications"`
	Status               string    `db:"status" json:"status"` // active, discontinued, draft
	CreatedAt            time.Time `db:"created_at" json:"created_at"`
	UpdatedAt            time.Time `db:"updated_at" json:"updated_at"`
	
	// Champs calculés
	CategoryName        string `db:"category_name" json:"category_name,omitempty"`
	DocumentationCount  int    `db:"documentation_count" json:"documentation_count,omitempty"`
	UserCount          int    `db:"user_count" json:"user_count,omitempty"`
}

// Category - Catégories de produits
type Category struct {
	ID           int       `db:"id" json:"id"`
	Name         string    `db:"name" json:"name"`
	Description  string    `db:"description" json:"description"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
	
	// Champs calculés
	ProductCount int `db:"product_count" json:"product_count,omitempty"`
}

// ProductDocument - Documents liés aux produits
type ProductDocument struct {
	ID          int       `db:"id" json:"id"`
	ProductID   int       `db:"product_id" json:"product_id"`
	Name        string    `db:"name" json:"name"`
	Description string    `db:"description" json:"description"`
	FileType    string    `db:"file_type" json:"file_type"` // manual, datasheet, warranty, image
	FilePath    string    `db:"file_path" json:"file_path"`
	FileSize    int64     `db:"file_size" json:"file_size"`
	UploadedAt  time.Time `db:"uploaded_at" json:"uploaded_at"`
}

// CreateProductRequest - Structure pour la création de produit
type CreateProductRequest struct {
	Name                string   `json:"name"`
	CategoryID          *int     `json:"category_id"`
	Brand               string   `json:"brand"`
	Model               string   `json:"model"`
	Description         string   `json:"description"`
	Price               *float64 `json:"price"`
	WarrantyMonths      int      `json:"warranty_months"`
	WarrantyConditions  string   `json:"warranty_conditions"`
	ManufacturerWebsite string   `json:"manufacturer_website"`
	Specifications      string   `json:"specifications"`
	Status              string   `json:"status"`
}