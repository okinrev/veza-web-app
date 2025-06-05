// file: backend/models/tag.go

package models

type Tag struct {
	ID   int    `db:"id" json:"id"`
	Name string `db:"name" json:"name"`
}