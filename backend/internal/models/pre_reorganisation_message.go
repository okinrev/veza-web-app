//file: backend/models/chat.go

package models

import "time"

type Message struct {
	ID        int       `db:"id" json:"id"`
	FromUser  int       `db:"from_user" json:"from_user"`
	ToUser    *int      `db:"to_user,omitempty" json:"to_user,omitempty"`
	Room      *string   `db:"room,omitempty" json:"room,omitempty"`
	Content   string    `db:"content" json:"content"`
	Timestamp time.Time `db:"timestamp" json:"timestamp"`
}
