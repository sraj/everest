package model

import "time"

// Document represents a document in the system.
type Document struct {
	ID          string    `json:"id" db:"id"`
	Title       string    `json:"title" db:"title"`
	OwnerID     string    `json:"owner_id" db:"owner_id"`
	ContentID   string    `json:"content_id" db:"content_id"`     // MinIO object key for content
	ThumbnailID *string   `json:"thumbnail_id" db:"thumbnail_id"` // MinIO object key for thumbnail
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}
