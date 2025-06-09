//file: backend/models/chat.go

package models

import "time"

type Room struct {
	ID        int       `db:"id" json:"id"`
	Name      string    `db:"name" json:"name"`
	IsPrivate bool      `db:"is_private" json:"is_private"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
