//file: backend/models/shared_ressource.go

package models

import "time"
import "github.com/lib/pq"

type SharedResource struct {
	ID               int             `db:"id" json:"id"`
	Title            string          `db:"title" json:"title"`
	Description 	 *string 		 `db:"description" json:"description,omitempty"`
	Filename         string          `db:"filename" json:"filename"`
	URL              string          `db:"url" json:"url"`
	Type             string          `db:"type" json:"type"`
	Tags             pq.StringArray `db:"tags" json:"tags"`
	UploaderID       int             `db:"uploader_id" json:"uploader_id"`
	UploaderUsername string          `db:"uploader_username" json:"uploader_username"`
	IsPublic         bool            `db:"is_public" json:"is_public"`
	UploadedAt       time.Time       `db:"uploaded_at" json:"uploaded_at"`
	DownloadCount    int             `db:"download_count" json:"download_count"`
}
