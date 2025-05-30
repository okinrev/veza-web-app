//file: backend/models/shared_ressource.go

package models

import "time"

type SharedResource struct {
	ID         int       `db:"id" json:"id"`
	Title      string    `db:"title" json:"title"`
	Filename   string    `db:"filename" json:"filename"`
	URL        string    `db:"url" json:"url"`
	Type       string    `db:"type" json:"type"`
	Tags       []string  `db:"tags" json:"tags"`
	UploaderID int       `db:"uploader_id" json:"uploader_id"`
	IsPublic   bool      `db:"is_public" json:"is_public"`
	UploadedAt time.Time `db:"uploaded_at" json:"uploaded_at"`
}
