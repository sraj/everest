package service

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"github.com/sraj/everest/internal/store"
)

// ChatConfig holds configuration for the chat service.
type ChatConfig struct {
	Endpoint string // Bifrost base URL
	Model    string
}

// DefaultChatConfig reads from env vars.
func DefaultChatConfig() ChatConfig {
	bifrostURL := strings.TrimRight(envOrDefault("BIFROST_ENDPOINT", "http://localhost:8081"), "/")
	return ChatConfig{
		Endpoint: bifrostURL + "/v1/chat/completions",
		Model:    envOrDefault("LLM_MODEL", "openrouter/nvidia/nemotron-3-ultra-550b-a55b:free"),
	}
}

// ChatService orchestrates RAG-based chat.
type ChatService interface {
	// Stream generates a response using RAG and writes SSE events to w.
	Stream(ctx context.Context, message string, w io.Writer) error
	// Process generates a response using RAG and returns the complete text.
	Process(ctx context.Context, message string) (string, []map[string]any, error)
}

type chatService struct {
	cfg       ChatConfig
	vector    store.VectorStore
	embedding EmbeddingService
	log       *slog.Logger
}

// NewChatService creates a new chat service.
func NewChatService(cfg ChatConfig, vector store.VectorStore, embedding EmbeddingService, log *slog.Logger) ChatService {
	return &chatService{
		cfg:       cfg,
		vector:    vector,
		embedding: embedding,
		log:       log,
	}
}

// Process generates a RAG response and returns the complete text with sources.
func (c *chatService) Process(ctx context.Context, message string) (string, []map[string]any, error) {
	var sources []map[string]any
	contextText := ""

	// Try RAG — if anything fails, fall back to direct LLM call.
	queryEmbedding, model, embErr := c.embedding.Generate(ctx, []byte("<p>"+message+"</p>"))
	if embErr != nil {
		c.log.Warn("embedding generation failed", "error", embErr)
	} else if len(queryEmbedding) == 0 {
		c.log.Warn("embedding generation returned empty vector")
	} else {
		c.log.Debug("embedding generated", "model", model, "dimensions", len(queryEmbedding))
		results, qErr := c.vector.Query(ctx, queryEmbedding, 5)
		if qErr != nil {
			c.log.Warn("vector query failed", "error", qErr)
		} else if len(results) == 0 {
			c.log.Warn("vector query returned no results")
		} else {
			contextText = c.buildContext(results)
			sources = c.buildSources(results)
			c.log.Debug("rag context built", "sources", len(sources), "contextLen", len(contextText))
		}
	}

	answer, err := c.complete(ctx, message, contextText)
	if err != nil {
		return "", nil, fmt.Errorf("llm complete: %w", err)
	}
	return answer, sources, nil
}

// complete calls the LLM and returns the full response (non-streaming).
func (c *chatService) complete(ctx context.Context, message, contextText string) (string, error) {
	userMessage := message
	if contextText != "" {
		userMessage = fmt.Sprintf("Use the following documents to answer the question. If the answer is not in the documents, say you don't have that information.\n\nDocuments:\n%s\n\nQuestion: %s", contextText, message)
	}

	body := map[string]any{
		"model":  c.cfg.Model,
		"stream": false,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant for Everest. Answer questions using the documents provided in the user message when available."},
			{"role": "user", "content": userMessage},
		},
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return "", fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.cfg.Endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return "", fmt.Errorf("create llm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("llm api returned %d: %s", resp.StatusCode, string(b))
	}

	// Read the full body so we can log it on failure.
	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read llm response: %w", err)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(rawBody, &result); err != nil {
		c.log.Error("failed to decode llm response", "error", err, "body", string(rawBody))
		return "", fmt.Errorf("decode llm response: %w", err)
	}

	if len(result.Choices) == 0 {
		c.log.Warn("llm returned empty choices", "body", string(rawBody))
		return "", fmt.Errorf("empty llm response")
	}
	return result.Choices[0].Message.Content, nil
}

// Stream generates a RAG response and writes SSE events.
func (c *chatService) Stream(ctx context.Context, message string, w io.Writer) error {
	// 1. Generate embedding for the query
	queryEmbedding, _, err := c.embedding.Generate(ctx, []byte("<p>"+message+"</p>"))
	if err != nil {
		c.log.Warn("embedding generation failed in stream", "error", err)
		return c.streamLLM(ctx, message, "", w)
	}
	if len(queryEmbedding) == 0 {
		c.log.Warn("embedding generation returned empty vector in stream")
		return c.streamLLM(ctx, message, "", w)
	}

	// 2. Search for relevant documents
	results, err := c.vector.Query(ctx, queryEmbedding, 5)
	if err != nil {
		c.log.Warn("vector query failed in stream", "error", err)
		return c.streamLLM(ctx, message, "", w)
	}

	// 3. Build context from search results
	contextText := c.buildContext(results)
	if contextText != "" {
		c.sendSSE(w, "citations", map[string]any{"sources": c.buildSources(results)})
	}

	// 4. Stream LLM response
	return c.streamLLM(ctx, message, contextText, w)
}

func (c *chatService) buildContext(results []store.QueryResult) string {
	if len(results) == 0 {
		return ""
	}
	var parts []string
	for i, r := range results {
		if r.Document == "" {
			continue
		}
		title := "Untitled"
		if t, ok := r.Metadata["title"]; ok && t != "" {
			title = t
		}
		cleaned := stripHTML([]byte(r.Document))
		parts = append(parts, fmt.Sprintf("Document %d (title: %s):\n%s", i+1, title, cleaned))
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n")
}

func (c *chatService) buildSources(results []store.QueryResult) []map[string]any {
	sources := make([]map[string]any, len(results))
	for i, r := range results {
		title := "Untitled"
		if t, ok := r.Metadata["title"]; ok && t != "" {
			title = t
		}
		sources[i] = map[string]any{
			"title":   title,
			"snippet": truncateText(stripHTML([]byte(r.Document)), 200),
			"score":   r.Distance,
		}
	}
	return sources
}

func (c *chatService) streamLLM(ctx context.Context, message, contextText string, w io.Writer) error {
	userMessage := message
	if contextText != "" {
		userMessage = fmt.Sprintf("Use the following documents to answer the question. If the answer is not in the documents, say you don't have that information.\n\nDocuments:\n%s\n\nQuestion: %s", contextText, message)
	}

	body := map[string]any{
		"model":  c.cfg.Model,
		"stream": true,
		"messages": []map[string]string{
			{"role": "system", "content": "You are a helpful assistant for Everest. Answer questions using the documents provided in the user message when available."},
			{"role": "user", "content": userMessage},
		},
	}

	reqBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.cfg.Endpoint, bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("create llm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("llm request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llm api returned %d: %s", resp.StatusCode, string(b))
	}

	// Stream SSE tokens
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			c.sendSSE(w, "token", chunk.Choices[0].Delta.Content)
		}
	}

	c.sendSSE(w, "done", map[string]string{"status": "complete"})
	return nil
}

func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

func (c *chatService) sendSSE(w io.Writer, event string, data any) {
	var payload string
	switch v := data.(type) {
	case string:
		payload = v
	default:
		b, _ := json.Marshal(v)
		payload = string(b)
	}
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, payload)
	if f, ok := w.(flusher); ok {
		f.Flush()
	}
}

type flusher interface {
	Flush() error
}
