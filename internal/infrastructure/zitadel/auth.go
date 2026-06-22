package zitadel

import (
	"context"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/zitadel/oidc/v3/pkg/client"
	"github.com/zitadel/oidc/v3/pkg/oidc"
)

// IntrospectUser holds the verified user info extracted from a token.
type IntrospectUser struct {
	Sub     string `json:"sub"`
	Email   string `json:"email,omitempty"`
	Name    string `json:"name,omitempty"`
	Picture string `json:"picture,omitempty"`
}

// Verifier validates Bearer tokens via OIDC introspection.
type Verifier struct {
	clientID     string
	introspectFn func(ctx context.Context, token string) (*oidc.IntrospectionResponse, error)
	log          *slog.Logger
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

	introspectFn := func(ctx context.Context, token string) (*oidc.IntrospectionResponse, error) {
		return client.Introspect[*oidc.IntrospectionResponse](ctx,
			discovery.IntrospectionEndpoint,
			clientID,
			"", // client secret not needed for public clients without auth
			token,
			httpClient,
		)
	}

	return &Verifier{clientID: clientID, introspectFn: introspectFn, log: log}, nil
}

// hostHeaderTransport overrides the Host header so Zitadel resolves the correct instance.
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
				"error": "invalid authorization scheme, use Bearer",
			})
		}

		resp, err := v.introspectFn(c.Context(), token)
		if err != nil {
			v.log.Error("token introspection failed", "error", err.Error())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token: " + err.Error(),
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

		resp, err := v.introspectFn(c.Context(), token)
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
