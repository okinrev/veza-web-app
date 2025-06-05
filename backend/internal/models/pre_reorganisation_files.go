//file: backend/models/file.go

package models

import "time"

type File struct {
	ID         int       `db:"id" json:"id"`
	ProductID  int       `db:"product_id" json:"product_id"`
	Filename   string    `db:"filename" json:"filename"`
	URL        string    `db:"url" json:"url"`
	Type       string    `db:"type" json:"type"` // exemple : "manual", "diagram"
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
}
