package model

import "time"

// DocumentPermission represents document sharing permissions.
type DocumentPermission struct {
	ID         string    `json:"id" db:"id"`
	DocumentID string    `json:"document_id" db:"document_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	Role       string    `json:"role" db:"role"` // owner, editor, viewer
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
