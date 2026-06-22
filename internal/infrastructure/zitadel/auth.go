package zitadel

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc/v3"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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

// customClaims extends jwt.RegisteredClaims with Zitadel-specific fields.
type customClaims struct {
	jwt.RegisteredClaims
	Email string `json:"email"`
	Name  string `json:"name"`
}

// Verifier validates Bearer tokens via JWKS.
type Verifier struct {
	jwks keyfunc.Keyfunc
	iss  string
	log  *slog.Logger
}

// NewVerifier creates a verifier that validates tokens using Zitadel's JWKS.
// No client credentials needed — suitable for public SPA clients.
func NewVerifier(issuer string, log *slog.Logger) (*Verifier, error) {
	httpClient := &http.Client{
		Transport: &hostHeaderTransport{base: http.DefaultTransport, host: "localhost:8082"},
		Timeout:   30 * time.Second,
	}

	// Use the internal Docker service URL for discovery while passing localhost
	// as the issuer (matches what Zitadel expects for instance resolution).
	wellKnown := issuer + "/.well-known/openid-configuration"
	discovery, err := client.Discover(context.Background(), "http://localhost:8082", httpClient, wellKnown)
	if err != nil {
		return nil, err
	}

	rawJWKS, err := fetchJWKS(httpClient, discovery, issuer)
	if err != nil {
		return nil, err
	}

	jwks, err := keyfunc.NewJWKSetJSON(rawJWKS)
	if err != nil {
		return nil, err
	}

	return &Verifier{jwks: jwks, iss: discovery.Issuer, log: log}, nil
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

// fetchJWKS fetches the JWKS JSON from Zitadel's JWKS URI.
// It rewrites the host in the URI so the request reaches the correct internal
// service (e.g., zitadel-proxy:80 instead of localhost:8082).
func fetchJWKS(httpClient *http.Client, d *oidc.DiscoveryConfiguration, internalIssuer string) (json.RawMessage, error) {
	jwksURI := d.JwksURI

	// Rewrite localhost to the internal service URL for Docker network access.
	if jwksParsed, err := url.Parse(jwksURI); err == nil {
		if strings.Contains(jwksParsed.Host, "localhost") {
			if internalParsed, err := url.Parse(internalIssuer); err == nil {
				jwksParsed.Host = internalParsed.Host
				jwksURI = jwksParsed.String()
			}
		}
	}

	resp, err := httpClient.Get(jwksURI)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var raw json.RawMessage
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		return nil, err
	}
	return raw, nil
}

// Middleware returns a Fiber handler that validates Bearer tokens and sets
// c.Locals("user") with the verified user info.
func (v *Verifier) Middleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		auth := c.Get("Authorization")
		if auth == "" {
			v.log.Warn("missing authorization header", "path", c.Path(), "method", c.Method())
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

		claims := &customClaims{}
		parsed, err := jwt.ParseWithClaims(token, claims, v.jwks.Keyfunc,
			jwt.WithIssuer(v.iss),
			jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
			jwt.WithLeeway(30*time.Second),
		)
		if err != nil {
			v.log.Error("token validation failed", "error", err.Error())
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "invalid token: " + err.Error(),
			})
		}
		if !parsed.Valid {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "token is not valid",
			})
		}

		user := IntrospectUser{
			Sub:   claims.Subject,
			Email: claims.Email,
			Name:  claims.Name,
		}
		c.Locals("user", user)
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

		claims := &customClaims{}
		parsed, err := jwt.ParseWithClaims(token, claims, v.jwks.Keyfunc,
			jwt.WithIssuer(v.iss),
			jwt.WithValidMethods([]string{"RS256", "RS384", "RS512"}),
			jwt.WithLeeway(30*time.Second),
		)
		if err != nil || !parsed.Valid {
			return c.Next()
		}

		c.Locals("user", IntrospectUser{
			Sub:   claims.Subject,
			Email: claims.Email,
			Name:  claims.Name,
		})
		return c.Next()
	}
}
