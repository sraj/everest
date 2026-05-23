package model

import "time"

// Document represents a document in the system
type Document struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	ContentID   string    `json:"content_id" db:"content_id"`     // MinIO object key for content
	ThumbnailID string    `json:"thumbnail_id" db:"thumbnail_id"` // MinIO object key for thumbnail
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// User represents a user in the system
type User struct {
	ID        string    `json:"id" db:"id"`
	Email     string    `json:"email" db:"email"`
	Name      string    `json:"name" db:"name"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// DocumentPermission represents document sharing permissions
type DocumentPermission struct {
	ID         string    `json:"id" db:"id"`
	DocumentID string    `json:"document_id" db:"document_id"`
	UserID     string    `json:"user_id" db:"user_id"`
	Role       string    `json:"role" db:"role"` // owner, editor, viewer
	CreatedAt  time.Time `json:"created_at" db:"created_at"`
}
