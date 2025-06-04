//file: backend/models/ressource.go

package models

import "time"

type InternalRessource struct {
	ID         int       `db:"id" json:"id"`
	ProductID  int       `db:"product_id" json:"product_id"`
	Title      string    `db:"title" json:"title"`
	Filename   string    `db:"filename" json:"filename"`
	URL        string    `db:"url" json:"url"`
	Type       string    `db:"type" json:"type"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
}
