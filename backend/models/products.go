//file: backend/models/products.go

package models

import "time"

type Product struct {
	ID              int       `db:"id" json:"id"`
	UserID          int       `db:"user_id" json:"-"`
	Name            string    `db:"name" json:"name"`
	Version         string    `db:"version" json:"version"`
	PurchaseDate    time.Time `db:"purchase_date" json:"purchase_date"`
	WarrantyExpires time.Time `db:"warranty_expires" json:"warranty_expires"`
}
