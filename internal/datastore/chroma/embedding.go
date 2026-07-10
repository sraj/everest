package chroma

import (
	"context"
	"fmt"
	"net/http"

	"github.com/sraj/everest/internal/store"
	"github.com/sraj/everest/pkg/httpclient"
)

// Config holds ChromaDB connection settings.
type Config struct {
	Endpoint   string
	Collection string
}

// vectorStore implements store.VectorStore backed by ChromaDB's REST API.
type vectorStore struct {
	client       *httpclient.Client
	collection   string // collection name
	collectionID string // resolved UUID for v2 API paths
}

// NewVectorStore creates a ChromaDB-backed VectorStore. It ensures the
// named collection exists, creating it if necessary.
func NewVectorStore(ctx context.Context, cfg Config) (store.VectorStore, error) {
	s := &vectorStore{
		client:     httpclient.New(cfg.Endpoint),
		collection: cfg.Collection,
	}

	fmt.Printf("ChromaDB Endpoint: %s, Collection: %s\n", cfg.Endpoint, cfg.Collection)
	if err := s.ensureCollection(ctx); err != nil {
		return nil, fmt.Errorf("chroma: ensure collection: %w", err)
	}

	return s, nil
}

const v2BasePath = "/api/v2/tenants/default_tenant/databases/default_database/collections"

// v2CollectionPath returns the base v2 API path for the configured collection UUID.
func (s *vectorStore) v2CollectionPath(suffix string) string {
	return v2BasePath + "/" + s.collectionID + suffix
}

// ensureCollection resolves the collection UUID, creating it if it doesn't already exist.
func (s *vectorStore) ensureCollection(ctx context.Context) error {
	id, err := s.getCollectionID(ctx, s.collection)
	if err == nil {
		s.collectionID = id
		return nil
	}

	var created struct {
		ID string `json:"id"`
	}
	if err := s.client.DoJSON(ctx, "POST", v2BasePath, map[string]any{
		"name":     s.collection,
		"metadata": map[string]string{"description": "Everest document embeddings"},
	}, &created, http.StatusOK, http.StatusCreated); err != nil {
		return fmt.Errorf("create collection: %w", err)
	}
	s.collectionID = created.ID
	return nil
}

// getCollectionID fetches a collection by name and returns its UUID.
func (s *vectorStore) getCollectionID(ctx context.Context, name string) (string, error) {
	var col struct {
		ID string `json:"id"`
	}
	if err := s.client.DoJSON(ctx, "GET", v2BasePath+"/"+name, nil, &col, http.StatusOK); err != nil {
		return "", err
	}
	return col.ID, nil
}

// Upsert inserts or updates an embedding vector with metadata.
func (s *vectorStore) Upsert(ctx context.Context, documentID string, content []byte, embedding []float64, metadata map[string]string) error {
	return s.client.DoOK(ctx, "POST", s.v2CollectionPath("/upsert"), map[string]any{
		"ids":        []string{documentID},
		"documents":  []string{string(content)},
		"embeddings": [][]float64{embedding},
		"metadatas":  []map[string]string{metadata},
	})
}

// GetByDocumentID retrieves the stored embedding for a document.
func (s *vectorStore) GetByDocumentID(ctx context.Context, documentID string) ([]float64, error) {
	var result struct {
		IDs        []string    `json:"ids"`
		Embeddings [][]float64 `json:"embeddings"`
	}
	if err := s.client.DoJSON(ctx, "POST", s.v2CollectionPath("/get"), map[string]any{
		"ids":     []string{documentID},
		"include": []string{"embeddings"},
	}, &result, http.StatusOK); err != nil {
		return nil, err
	}

	if len(result.Embeddings) == 0 {
		return nil, store.ErrNotFound{Resource: "embedding", ID: documentID}
	}
	return result.Embeddings[0], nil
}

// Query searches for documents similar to the given embedding vector.
func (s *vectorStore) Query(ctx context.Context, embedding []float64, topK int) ([]store.QueryResult, error) {
	if topK <= 0 {
		topK = 5
	}

	var result struct {
		IDs        [][]string            `json:"ids"`
		Documents  [][]string            `json:"documents"`
		Metadatas  [][]map[string]string `json:"metadatas"`
		Distances  [][]float64           `json:"distances"`
	}
	if err := s.client.DoJSON(ctx, "POST", s.v2CollectionPath("/query"), map[string]any{
		"query_embeddings": [][]float64{embedding},
		"n_results":        topK,
		"include":          []string{"documents", "metadatas", "distances"},
	}, &result, http.StatusOK); err != nil {
		return nil, fmt.Errorf("chroma query: %w", err)
	}

	if len(result.IDs) == 0 || len(result.IDs[0]) == 0 {
		return nil, nil
	}

	batch := result.IDs[0]
	out := make([]store.QueryResult, len(batch))
	for i, id := range batch {
		qr := store.QueryResult{ID: id}
		if i < len(result.Documents[0]) {
			qr.Document = result.Documents[0][i]
		}
		if i < len(result.Metadatas[0]) {
			qr.Metadata = result.Metadatas[0][i]
		}
		if i < len(result.Distances[0]) {
			qr.Distance = result.Distances[0][i]
		}
		out[i] = qr
	}
	return out, nil
}

// Delete removes the embedding for a document.
func (s *vectorStore) Delete(ctx context.Context, documentID string) error {
	return s.client.DoOK(ctx, "POST", s.v2CollectionPath("/delete"), map[string]any{
		"ids": []string{documentID},
	})
}

// Close is a no-op for the ChromaDB HTTP client.
func (s *vectorStore) Close() error { return nil }

// Ping checks connectivity to the ChromaDB server.
func (s *vectorStore) Ping(ctx context.Context) error {
	return s.client.DoOK(ctx, "GET", "/api/v1/heartbeat", nil, http.StatusOK)
}
