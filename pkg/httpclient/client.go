package httpclient

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is a JSON HTTP client for talking to REST APIs.
type Client struct {
	BaseURL string
	HTTP    *http.Client
}

// New creates a Client with sensible defaults.
func New(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 30 * time.Second},
	}
}

// Do sends a JSON request and returns the raw response.
func (c *Client) Do(ctx context.Context, method, path string, body any) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request: %w", err)
		}
		bodyReader = bytes.NewReader(b)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.BaseURL+path, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	return c.HTTP.Do(req)
}

// DoOK sends a JSON request and returns an error if the status is not in allowed.
func (c *Client) DoOK(ctx context.Context, method, path string, body any, allowed ...int) error {
	resp, err := c.Do(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if statusOK(resp.StatusCode, allowed) {
		return nil
	}

	b, _ := io.ReadAll(resp.Body)
	return fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, string(b))
}

// DoJSON sends a JSON request, checks status, and unmarshals the response into dest.
func (c *Client) DoJSON(ctx context.Context, method, path string, body any, dest any, allowed ...int) error {
	resp, err := c.Do(ctx, method, path, body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if !statusOK(resp.StatusCode, allowed) {
		return fmt.Errorf("%s %s returned %d: %s", method, path, resp.StatusCode, string(b))
	}

	if dest != nil && len(b) > 0 {
		if err := json.Unmarshal(b, dest); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}

func statusOK(code int, allowed []int) bool {
	if len(allowed) == 0 {
		return code == http.StatusOK
	}
	for _, c := range allowed {
		if code == c {
			return true
		}
	}
	return false
}
