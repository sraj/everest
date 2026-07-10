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

	"github.com/sraj/everest/internal/domain/model"
)

type TaggerConfig struct {
	Endpoint string
	Model    string
	Enabled  bool
}

func DefaultTaggerConfig() TaggerConfig {
	bifrostURL := strings.TrimRight(envOrDefault("BIFROST_ENDPOINT", "http://localhost:8081"), "/")
	return TaggerConfig{
		Endpoint: bifrostURL + "/v1/chat/completions",
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

// TaggerService generates tags for document content using an AI model.
type TaggerService interface {
	Generate(ctx context.Context, content []byte) (model.Tags, error)
	Close()
}

type taggerService struct {
	cfg TaggerConfig
	log *slog.Logger
}

// NewTagger creates a tag generator backed by OpenRouter via Bifrost.
func NewTagger(cfg TaggerConfig, log *slog.Logger) TaggerService {
	return &taggerService{cfg: cfg, log: log}
}

// Generate calls the AI model and returns extracted tags.
func (t *taggerService) Generate(ctx context.Context, content []byte) (model.Tags, error) {
	if !t.cfg.Enabled {
		return nil, nil
	}

	text := stripHTML(content)
	if len(text) < 50 {
		return nil, nil
	}
	text = text[:min(len(text), 3000)]

	return t.callAI(ctx, text)
}

func (t *taggerService) callAI(ctx context.Context, text string) (model.Tags, error) {
	systemPrompt := `Extract 3-5 relevant tags from the document content. Return a JSON object with a "tags" array. Example: {"tags":["technology","programming","golang"]}. No other text.`

	reqBody := map[string]any{
		"model": t.cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": "Document content:\n\n" + text},
		},
		"temperature":     0.3,
		"max_completion_tokens": 500,
		"seed":             42,
		"response_format":  map[string]string{"type": "json_object"},
		"reasoning":        map[string]any{"effort": "low"},
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

func (t *taggerService) Close() {}

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
