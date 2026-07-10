package store

import (
	"context"
)

// QueryResult holds a single match from a vector search.
type QueryResult struct {
	ID       string
	Document string
	Metadata map[string]string
	Distance float64
}

// VectorStore defines the interface for vector embedding persistence (ChromaDB-backed).
type VectorStore interface {
	// Upsert creates or updates the embedding for a document. The embedding vector
	// is stored alongside metadata (title, owner_id) for future semantic search.
	Upsert(ctx context.Context, documentID string, content []byte, embedding []float64, metadata map[string]string) error
	// GetByDocumentID retrieves the stored embedding for a document.
	GetByDocumentID(ctx context.Context, documentID string) ([]float64, error)
	// Query searches for documents similar to the given embedding vector.
	Query(ctx context.Context, embedding []float64, topK int) ([]QueryResult, error)
	// Delete removes the embedding for a document.
	Delete(ctx context.Context, documentID string) error
	// Close cleans up any resources.
	Close() error
}
