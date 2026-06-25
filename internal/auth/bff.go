package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	gorilla "github.com/gorilla/securecookie"
)

const (
	cookieName       = "everest_session"
	cookieMaxAge     = 8 * 60 * 60 // 8 hours
	pkceCookie       = "everest_pkce"
	pkceCookieMaxAge = 5 * 60 // 5 minutes
)

// ── Session cookie (gorilla/securecookie) ────────────────────────

type sessionCookie struct {
	sc  *gorilla.SecureCookie
	log *slog.Logger
}

func newSessionCookie(secret string, log *slog.Logger) *sessionCookie {
	key := []byte(secret)
	if secret == "" {
		key = make([]byte, 32)
		rand.Read(key)
		log.Warn("no SESSION_SECRET set, using random key — sessions lost on restart")
	}
	return &sessionCookie{
		sc:  gorilla.New(key),
		log: log,
	}
}

func (s *sessionCookie) set(c *fiber.Ctx, value any) error {
	encoded, err := s.sc.Encode(cookieName, value)
	if err != nil {
		return fmt.Errorf("encode cookie: %w", err)
	}
	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    encoded,
		Domain:   cookieDomain(c),
		Path:     "/",
		Expires:  time.Now().Add(cookieMaxAge),
		HTTPOnly: true,
		Secure:   isSecure(c),
		SameSite: "Lax",
	})
	return nil
}

func (s *sessionCookie) clear(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HTTPOnly: true,
		SameSite: "Lax",
	})
}

func (s *sessionCookie) get(c *fiber.Ctx, dest any) error {
	val := c.Cookies(cookieName)
	if val == "" {
		return fmt.Errorf("no session cookie")
	}
	if err := s.sc.Decode(cookieName, val, dest); err != nil {
		return fmt.Errorf("invalid session: %w", err)
	}
	return nil
}

func (s *sessionCookie) setPKCE(c *fiber.Ctx, value any) error {
	encoded, err := s.sc.Encode(pkceCookie, value)
	if err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{
		Name:     pkceCookie,
		Value:    encoded,
		Path:     "/auth",
		Expires:  time.Now().Add(pkceCookieMaxAge * time.Second),
		HTTPOnly: true,
		Secure:   isSecure(c),
		SameSite: "Lax",
	})
	return nil
}

func (s *sessionCookie) getPKCE(c *fiber.Ctx, dest any) error {
	val := c.Cookies(pkceCookie)
	if val == "" {
		return fmt.Errorf("no pkce cookie")
	}
	return s.sc.Decode(pkceCookie, val, dest)
}

func (s *sessionCookie) clearPKCE(c *fiber.Ctx) {
	c.Cookie(&fiber.Cookie{
		Name:     pkceCookie,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HTTPOnly: true,
	})
}

func isSecure(c *fiber.Ctx) bool {
	return strings.HasPrefix(c.Protocol(), "https")
}

// cookieDomain returns the domain for cookies so they are shared across
// ports on localhost. In production (non-localhost), it returns empty
// (scope to the exact host, which is the right choice behind a proxy).
func cookieDomain(c *fiber.Ctx) string {
	host := c.Hostname()
	if host == "localhost" || host == "127.0.0.1" {
		return "localhost"
	}
	return ""
}

// ── OIDC BFF Handler ───────────────────────────────────────────

type BFFConfig struct {
	Issuer        string
	ClientID      string
	RedirectURI   string
	PostLogoutURI string
	SessionSecret string
	Log           *slog.Logger
}

type BFFHandler struct {
	cfg       BFFConfig
	cookie    *sessionCookie
	client    *http.Client
	endpoints oidcEndpoints
	log       *slog.Logger
}

type oidcEndpoints struct {
	auth       string
	token      string
	endSession string
}

func NewBFFHandler(cfg BFFConfig) (*BFFHandler, error) {
	if cfg.ClientID == "" {
		return nil, fmt.Errorf("ZITADEL_CLIENT_ID is required for BFF auth")
	}
	if cfg.RedirectURI == "" {
		cfg.RedirectURI = "http://localhost:8080/auth/callback"
	}
	if cfg.PostLogoutURI == "" {
		cfg.PostLogoutURI = "/"
	}

	httpClient := &http.Client{Timeout: 30 * time.Second}

	wellKnown := strings.TrimRight(cfg.Issuer, "/") + "/.well-known/openid-configuration"
	resp, err := httpClient.Get(wellKnown)
	if err != nil {
		return nil, fmt.Errorf("oidc discovery: %w", err)
	}
	defer resp.Body.Close()

	var disc discoveryConfig
	if err := json.NewDecoder(resp.Body).Decode(&disc); err != nil {
		return nil, fmt.Errorf("parse discovery: %w", err)
	}

	return &BFFHandler{
		cfg:    cfg,
		cookie: newSessionCookie(cfg.SessionSecret, cfg.Log),
		client: httpClient,
		endpoints: oidcEndpoints{
			auth:       disc.AuthorizationEndpoint,
			token:      disc.TokenEndpoint,
			endSession: disc.EndSessionEndpoint,
		},
		log: cfg.Log,
	}, nil
}

// RegisterRoutes adds BFF auth routes.
func (h *BFFHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/auth/login", h.handleLogin)
	app.Get("/auth/callback", h.handleCallback)
	app.Get("/auth/logout", h.handleLogout)
	app.Get("/auth/me", h.handleMe)
}

func (h *BFFHandler) handleLogin(c *fiber.Ctx) error {
	state := randomString(32)
	codeVerifier := randomString(64)

	pkce := map[string]string{"state": state, "code_verifier": codeVerifier}
	if err := h.cookie.setPKCE(c, pkce); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to store login state",
		})
	}

	codeChallenge := base64.RawURLEncoding.EncodeToString(sha256Sum([]byte(codeVerifier)))

	authURL, _ := url.Parse(h.endpoints.auth)
	q := authURL.Query()
	q.Set("client_id", h.cfg.ClientID)
	q.Set("redirect_uri", h.cfg.RedirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "openid profile email")
	q.Set("state", state)
	q.Set("code_challenge", codeChallenge)
	q.Set("code_challenge_method", "S256")
	authURL.RawQuery = q.Encode()

	return c.Redirect(authURL.String(), http.StatusFound)
}

func (h *BFFHandler) handleCallback(c *fiber.Ctx) error {
	queryState := c.Query("state")
	code := c.Query("code")
	if queryState == "" || code == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "missing state or code",
		})
	}

	var pkce map[string]string
	if err := h.cookie.getPKCE(c, &pkce); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid or expired login session",
		})
	}
	if pkce["state"] != queryState {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "state mismatch",
		})
	}
	h.cookie.clearPKCE(c)

	tokenForm := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {h.cfg.RedirectURI},
		"client_id":     {h.cfg.ClientID},
		"code_verifier": {pkce["code_verifier"]},
	}

	tokenReq, _ := http.NewRequestWithContext(c.Context(), "POST",
		h.endpoints.token, strings.NewReader(tokenForm.Encode()))
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := h.client.Do(tokenReq)
	if err != nil {
		h.log.Error("token exchange failed", "error", err.Error())
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "authentication failed",
		})
	}
	defer tokenResp.Body.Close()

	if tokenResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(tokenResp.Body)
		h.log.Error("token exchange non-200", "status", tokenResp.StatusCode, "body", string(body))
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "token exchange failed",
		})
	}

	var tokenResult tokenResponse
	if err := json.NewDecoder(tokenResp.Body).Decode(&tokenResult); err != nil {
		h.log.Error("decode token response", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to parse token response",
		})
	}

	user, err := parseIDTokenClaims(tokenResult.IDToken)
	if err != nil {
		h.log.Error("parse id token", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "failed to parse identity",
		})
	}

	if err := h.cookie.set(c, *user); err != nil {
		h.log.Error("set session cookie", "error", err.Error())
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "session creation failed",
		})
	}

	h.log.Info("user authenticated via OIDC", "sub", user.Sub, "email", user.Email)
	return c.Redirect(h.cfg.PostLogoutURI, http.StatusFound)
}

func (h *BFFHandler) handleLogout(c *fiber.Ctx) error {
	h.cookie.clear(c)
	h.cookie.clearPKCE(c)
	return c.Redirect(h.cfg.PostLogoutURI, http.StatusFound)
}

func (h *BFFHandler) handleMe(c *fiber.Ctx) error {
	var user IntrospectUser
	if err := h.cookie.get(c, &user); err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "not authenticated",
		})
	}
	return c.JSON(user)
}

// BFFMiddleware validates the session cookie and sets c.Locals("user").
func (h *BFFHandler) BFFMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		var user IntrospectUser
		if err := h.cookie.get(c, &user); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "not authenticated",
			})
		}
		c.Locals("user", user)
		return c.Next()
	}
}

// ── Helpers ─────────────────────────────────────────────────────

func randomString(n int) string {
	b := make([]byte, n)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func sha256Sum(data []byte) []byte {
	h := sha256.Sum256(data)
	return h[:]
}

func parseIDTokenClaims(raw string) (*IntrospectUser, error) {
	parts := strings.Split(raw, ".")
	if len(parts) < 2 {
		return nil, fmt.Errorf("invalid JWT format")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, fmt.Errorf("decode JWT payload: %w", err)
	}
	var claims struct {
		Sub   string `json:"sub"`
		Email string `json:"email"`
		Name  string `json:"name"`
	}
	if err := json.Unmarshal(payload, &claims); err != nil {
		return nil, fmt.Errorf("parse JWT claims: %w", err)
	}
	return &IntrospectUser{
		Sub:   claims.Sub,
		Email: claims.Email,
		Name:  claims.Name,
	}, nil
}

type discoveryConfig struct {
	AuthorizationEndpoint string `json:"authorization_endpoint"`
	TokenEndpoint         string `json:"token_endpoint"`
	EndSessionEndpoint    string `json:"end_session_endpoint"`
}

type tokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
	RefreshToken string `json:"refresh_token"`
}
