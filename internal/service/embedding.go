package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

// EmbeddingConfig holds configuration for the embedding service.
type EmbeddingConfig struct {
	Endpoint string
	Model    string
	Enabled  bool
}

// DefaultEmbeddingConfig reads configuration from environment variables.
func DefaultEmbeddingConfig() EmbeddingConfig {
	bifrostURL := strings.TrimRight(envOrDefault("BIFROST_ENDPOINT", "http://localhost:8081"), "/")
	return EmbeddingConfig{
		Endpoint: bifrostURL + "/v1/embeddings",
		Model:    envOrDefault("EMBEDDING_MODEL", "openrouter/openai/text-embedding-3-small"),
		Enabled:  strings.ToLower(os.Getenv("EMBEDDING_ENABLED")) == "true",
	}
}

// EmbeddingService generates vector embeddings for document content.
type EmbeddingService interface {
	// Generate returns the embedding vector and the model used.
	Generate(ctx context.Context, content []byte) ([]float64, string, error)
	Close()
}

type embeddingService struct {
	cfg EmbeddingConfig
	log *slog.Logger
}

// NewEmbedding creates an embedding generator backed by OpenRouter via Bifrost.
func NewEmbedding(cfg EmbeddingConfig, log *slog.Logger) EmbeddingService {
	return &embeddingService{cfg: cfg, log: log}
}

// Generate calls the embedding model and returns the vector.
func (e *embeddingService) Generate(ctx context.Context, content []byte) ([]float64, string, error) {
	if !e.cfg.Enabled {
		return nil, "", nil
	}

	text := stripHTML(content)
	if len(text) < 10 {
		return nil, "", nil
	}

	return e.callAI(ctx, text)
}

func (e *embeddingService) callAI(ctx context.Context, text string) ([]float64, string, error) {
	reqBody := map[string]any{
		"model": e.cfg.Model,
		"input": text,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", e.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("embedding request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, "", fmt.Errorf("embedding api returned %d: %s", resp.StatusCode, string(b))
	}

	var embResp embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&embResp); err != nil {
		return nil, "", fmt.Errorf("decode embedding response: %w", err)
	}

	if len(embResp.Data) == 0 || len(embResp.Data[0].Embedding) == 0 {
		return nil, "", fmt.Errorf("empty embedding in response")
	}

	vec := embResp.Data[0].Embedding
	modelUsed := embResp.Model
	if modelUsed == "" {
		modelUsed = e.cfg.Model
	}

	return vec, modelUsed, nil
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
}

func (e *embeddingService) Close() {}
