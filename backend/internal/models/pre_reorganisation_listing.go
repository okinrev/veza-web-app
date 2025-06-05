//file: backend/models/listing.go

package models

type Listing struct {
    ID           int      `db:"id" json:"id"`
    UserID       int      `db:"user_id" json:"user_id"`
    ProductID    int      `db:"product_id" json:"product_id"`
    Description  string   `db:"description" json:"description"`
    State        string   `db:"state" json:"state"`
    Price        *int     `db:"price" json:"price,omitempty"`
    ExchangeFor  *string  `db:"exchange_for" json:"exchange_for,omitempty"`
    Images       []string `db:"images" json:"images"`
    Status       string   `db:"status" json:"status"`
    CreatedAt    string   `db:"created_at" json:"created_at"`
}