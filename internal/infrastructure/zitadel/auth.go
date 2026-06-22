package zitadel

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

// IntrospectUser holds the verified user info extracted from a token.
type IntrospectUser struct {
	Sub   string `json:"sub"`
	Email string `json:"email,omitempty"`
	Name  string `json:"name,omitempty"`
}

// Verifier validates Bearer tokens via OIDC introspection.
type Verifier struct {
	introspectionURL string
	clientID         string
	httpClient       *http.Client
	log              *slog.Logger
}

// NewVerifier creates a verifier that validates tokens using OIDC introspection.
func NewVerifier(issuer, clientID string, log *slog.Logger) (*Verifier, error) {
	httpClient := &http.Client{
		Transport: &hostHeaderTransport{base: http.DefaultTransport, host: "localhost:8082"},
		Timeout:   30 * time.Second,
	}

	wellKnown := issuer + "/.well-known/openid-configuration"
	discovery, err := client.Discover(context.Background(), "http://localhost:8082", httpClient, wellKnown)
	if err != nil {
		return nil, err
	}

	return &Verifier{
		introspectionURL: discovery.IntrospectionEndpoint,
		clientID:         clientID,
		httpClient:       httpClient,
		log:              log,
	}, nil
}

type hostHeaderTransport struct {
	base http.RoundTripper
	host string
}

func (t *hostHeaderTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Host = t.host
	return t.base.RoundTrip(req)
}

// Middleware returns a Fiber handler that validates Bearer tokens via introspection.
func (v *Verifier) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid authorization scheme",
			})
		}

		resp, err := v.introspect(c.Context(), token)
		if err != nil {
			v.log.Error("token introspection failed", "error", err.Error())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token",
			})
		}

		if !resp.Active {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "token is not active",
			})
		}

		c.Locals("user", IntrospectUser{
			Sub:   resp.Subject,
			Email: resp.Email,
			Name:  resp.Name,
		})
		return c.Next()
	}
}

// Optional is like Middleware but does not reject unauthenticated requests.
func (v *Verifier) Optional() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			return c.Next()
		}

		token, ok := strings.CutPrefix(auth, "Bearer ")
		if !ok {
			return c.Next()
		}

		resp, err := v.introspect(c.Context(), token)
		if err != nil || !resp.Active {
			return c.Next()
		}

		c.Locals("user", IntrospectUser{
			Sub:   resp.Subject,
			Email: resp.Email,
			Name:  resp.Name,
		})
		return c.Next()
	}
}

func (v *Verifier) introspect(ctx any, token string) (*oidc.IntrospectionResponse, error) {
	data := url.Values{}
	data.Set("token", token)
	data.Set("token_type_hint", "access_token")
	data.Set("client_id", v.clientID)

	req, err := http.NewRequestWithContext(context.Background(), "POST", v.introspectionURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Host = "localhost:8082"

	resp, err := v.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var introspectResp oidc.IntrospectionResponse
	if err := json.Unmarshal(body, &introspectResp); err != nil {
		return nil, err
	}

	return &introspectResp, nil
}
