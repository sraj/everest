package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	bifrost "github.com/maximhq/bifrost/core"
	"github.com/maximhq/bifrost/core/schemas"

	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/store"
)

// TaggerConfig holds AI tag generation configuration.
type TaggerConfig struct {
	Model   string // "openai/gpt-4o-mini", "ollama/llama3.2", "anthropic/claude-haiku"
	Enabled bool
}

// DefaultTaggerConfig returns sensible defaults.
func DefaultTaggerConfig() TaggerConfig {
	return TaggerConfig{
		Model:   "openai/gpt-4o-mini",
		Enabled: false,
	}
}

type tagger struct {
	cfg    TaggerConfig
	store  store.Store
	client *bifrost.Client
	log    *slog.Logger
}

// NewTagger creates a tag generator backed by the Bifrost Go SDK.
func NewTagger(cfg TaggerConfig, st store.Store, log *slog.Logger) (*tagger, error) {
	if !cfg.Enabled {
		return &tagger{cfg: cfg, store: st, log: log}, nil
	}

	client, err := bifrost.Init(context.Background(), schemas.BifrostConfig{
		Account: &tagAccount{},
	})
	if err != nil {
		return nil, fmt.Errorf("bifrost init: %w", err)
	}

	return &tagger{cfg: cfg, store: st, client: client, log: log}, nil
}

// GenerateAndSaveTags calls the AI model, extracts tags, and updates the document.
func (t *tagger) GenerateAndSaveTags(ctx context.Context, docID string, content []byte) {
	if !t.cfg.Enabled || t.client == nil {
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
	provider, modelName := parseModel(t.cfg.Model)

	systemPrompt := "Extract 3-5 relevant tags from the document content. Return ONLY a valid JSON array of lowercase strings. Example: [\"technology\",\"programming\",\"golang\"]. No other text."

	messages := []schemas.ChatMessage{
		{
			Role: schemas.ChatMessageRoleSystem,
			Content: &schemas.ChatMessageContent{
				ContentStr: &systemPrompt,
			},
		},
		{
			Role: schemas.ChatMessageRoleUser,
			Content: &schemas.ChatMessageContent{
				ContentStr: schemas.Ptr("Document content:\n\n" + text),
			},
		},
	}

	resp, err := t.client.ChatCompletionRequest(
		schemas.NewBifrostContext(ctx, schemas.NoDeadline),
		&schemas.BifrostChatRequest{
			Provider:    provider,
			Model:       modelName,
			Input:       messages,
			Temperature: schemas.Ptr(0.3),
			MaxTokens:   schemas.Ptr(200),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("ai request: %w", err)
	}

	if resp.Choices == nil || len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	raw := strings.TrimSpace(*resp.Choices[0].Message.Content.ContentStr)
	raw = strings.TrimPrefix(raw, "```json")
	raw = strings.TrimPrefix(raw, "```")
	raw = strings.TrimSuffix(raw, "```")
	raw = strings.TrimSpace(raw)

	var tags model.Tags
	if err := json.Unmarshal([]byte(raw), &tags); err != nil {
		t.log.Error("failed to parse tags from AI", "raw", raw, "error", err.Error())
		return nil, err
	}
	if len(tags) > 5 {
		tags = tags[:5]
	}
	return tags, nil
}

// parseModel splits "provider/model" into (provider, model) components.
func parseModel(modelArg string) (schemas.ModelProvider, string) {
	parts := strings.SplitN(modelArg, "/", 2)
	if len(parts) != 2 {
		return schemas.OpenAI, modelArg
	}
	switch strings.ToLower(parts[0]) {
	case "openai":
		return schemas.OpenAI, parts[1]
	case "anthropic":
		return schemas.Anthropic, parts[1]
	case "ollama":
		return schemas.Ollama, parts[1]
	case "google", "vertex":
		return schemas.GoogleVertexAI, parts[1]
	case "groq":
		return schemas.Groq, parts[1]
	case "mistral":
		return schemas.MistralAI, parts[1]
	default:
		return schemas.OpenAI, parts[1]
	}
}

// tagAccount implements schemas.Account for Bifrost initialization.
// API keys are read from environment variables (OPENAI_API_KEY, etc.).
type tagAccount struct{}

func (a *tagAccount) GetConfiguredProviders() ([]schemas.ModelProvider, error) {
	providers := []schemas.ModelProvider{schemas.OpenAI}
	if os.Getenv("ANTHROPIC_API_KEY") != "" {
		providers = append(providers, schemas.Anthropic)
	}
	return providers, nil
}

func (a *tagAccount) GetKeysForProvider(_ *context.Context, provider schemas.ModelProvider) ([]schemas.Key, error) {
	var key string
	switch provider {
	case schemas.OpenAI:
		key = os.Getenv("OPENAI_API_KEY")
	case schemas.Anthropic:
		key = os.Getenv("ANTHROPIC_API_KEY")
	}
	if key == "" {
		return nil, fmt.Errorf("no API key for %s", provider)
	}
	return []schemas.Key{{Value: key, Models: schemas.WhiteList{"*"}, Weight: 1.0}}, nil
}

func (a *tagAccount) GetConfigForProvider(provider schemas.ModelProvider) (*schemas.ProviderConfig, error) {
	return &schemas.ProviderConfig{
		NetworkConfig:            schemas.DefaultNetworkConfig,
		ConcurrencyAndBufferSize: schemas.DefaultConcurrencyAndBufferSize,
	}, nil
}

// Close shuts down the Bifrost client.
func (t *tagger) Close() {
	if t.client != nil {
		t.client.Shutdown()
	}
}
