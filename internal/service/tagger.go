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
	"time"

	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/store"
)

// TaggerConfig holds AI tag generation configuration.
type TaggerConfig struct {
	Endpoint string // Bifrost gateway URL (http://localhost:8081/v1/chat/completions)
	Model    string // "openrouter/nvidia/nemotron-3-ultra"
	Enabled  bool
}

// DefaultTaggerConfig returns sensible defaults.
func DefaultTaggerConfig() TaggerConfig {
	return TaggerConfig{
		Endpoint: envOrDefault("AI_TAGGER_ENDPOINT", "http://localhost:8081/v1/chat/completions"),
		Model:    envOrDefault("AI_TAGGER_MODEL", "openrouter/nvidia/nemotron-3-ultra-550b-a55b:free"),
		Enabled:  strings.ToLower(os.Getenv("AI_TAGGER_ENABLED")) == "true",
	}
}

func envOrDefault(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

type tagger struct {
	cfg   TaggerConfig
	store store.Store
	log   *slog.Logger
}

// NewTagger creates a tag generator that calls Bifrost (OpenAI-compatible API).
func NewTagger(cfg TaggerConfig, st store.Store, log *slog.Logger) *tagger {
	return &tagger{cfg: cfg, store: st, log: log}
}

// GenerateAndSaveTags calls the AI model, extracts tags, and updates the document.
func (t *tagger) GenerateAndSaveTags(ctx context.Context, docID string, content []byte) {
	if !t.cfg.Enabled {
		return
	}

	text := stripHTML(content)
	if len(text) < 50 {
		return
	}
	text = text[:min(len(text), 3000)]

	tags, err := t.callAI(ctx, text)
	if err != nil {
		t.log.Error("failed to generate tags", "error", err.Error(), "doc_id", docID)
		return
	}
	if len(tags) == 0 {
		return
	}

	doc, err := t.store.Document().GetByID(ctx, docID)
	if err != nil {
		t.log.Error("document not found for tags", "error", err.Error(), "doc_id", docID)
		return
	}

	doc.Tags = tags
	doc.UpdatedAt = time.Now()
	if err := t.store.Document().Update(ctx, doc); err != nil {
		t.log.Error("failed to save tags", "error", err.Error(), "doc_id", docID)
		return
	}

	t.log.Info("tags generated", "doc_id", docID, "tags", tags)
}

func (t *tagger) callAI(ctx context.Context, text string) (model.Tags, error) {
	systemPrompt := `Extract 3-5 relevant tags from the document content. Return a JSON object with a "tags" array. Example: {"tags":["technology","programming","golang"]}. No other text.`

	reqBody := map[string]any{
		"model": t.cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "Document content:\n\n" + text},
		},
		"temperature":     0.3,
		"max_completion_tokens": 200,
		"seed":             42,
		"response_format":  map[string]string{"type": "json_object"},
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(ctx, "POST", t.cfg.Endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("ai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ai returned %d: %s", resp.StatusCode, string(b))
	}

	var chatResp chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("decode ai response: %w", err)
	}

	if len(chatResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in ai response")
	}

	raw := strings.TrimSpace(chatResp.Choices[0].Message.Content)

	// With response_format: json_object, the output is a JSON object with a "tags" key.
	var result struct {
		Tags model.Tags `json:"tags"`
	}
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		t.log.Error("failed to parse tags from AI", "raw", raw, "error", err.Error())
		return nil, err
	}
	tags := result.Tags
	if len(tags) > 5 {
		tags = tags[:5]
	}
	return tags, nil
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func stripHTML(content []byte) string {
	s := string(content)
	for {
		start := strings.Index(s, "<")
		if start < 0 {
			break
		}
		end := strings.Index(s[start:], ">")
		if end < 0 {
			break
		}
		end += start
		s = s[:start] + s[end+1:]
	}
	return strings.TrimSpace(s)
}

// Close is a no-op for the HTTP-based tagger.
func (t *tagger) Close() {}

// InitBifrostProvider configures the OpenRouter provider in Bifrost via its REST API.
// Call once at startup if BIFROST_URL is set.
func InitBifrostProvider(ctx context.Context, bifrostURL, openRouterKey string) error {
	if bifrostURL == "" || openRouterKey == "" {
		return nil
	}
	payload := map[string]any{
		"provider": "openrouter",
		"keys": []map[string]any{
			{"name": "or-key", "value": openRouterKey, "models": []string{"*"}, "weight": 1.0},
		},
	}
	body, _ := json.Marshal(payload)
	req, err := http.NewRequestWithContext(ctx, "POST", bifrostURL+"/api/providers", bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("bifrost init: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("bifrost init failed (%d): %s", resp.StatusCode, string(b))
	}
	return nil
}
