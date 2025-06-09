//file: backend/models/tracks.go

package models

import (
	"time"
	"github.com/lib/pq"
)

type Track struct {
	ID              int       		`db:"id" json:"id"`
	Title           string    		`db:"title" json:"title"`
	Artist          string    		`db:"artist" json:"artist"`
	Filename        string    		`db:"filename" json:"filename"`
	CreatedAt       time.Time 		`db:"created_at" json:"created_at"`
	DurationSeconds *int       		`db:"duration_seconds" json:"duration_seconds"`
	Tags            pq.StringArray  `db:"tags" json:"tags"`
	IsPublic        *bool      		`db:"is_public" json:"is_public"`
	UploaderID      int       		`db:"uploader_id" json:"uploader_id"`
}
