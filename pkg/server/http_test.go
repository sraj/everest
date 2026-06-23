package server

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

type testRoutes struct {
	fn func(app *fiber.App)
}

func (r testRoutes) RegisterRoutes(app *fiber.App) {
	r.fn(app)
}

func TestNewHTTP_Routes(t *testing.T) {
	cfg := HTTPConfig{
		AppName: "test",
		Port:    "8080",
	}
	routes := testRoutes{fn: func(app *fiber.App) {
		app.Get("/hello", func(c *fiber.Ctx) error {
			return c.JSON(fiber.Map{"msg": "world"})
		})
	}}

	srv := NewHTTP(cfg, routes, testLog())
	httpSrv, ok := srv.(*HTTPServer)
	if !ok {
		t.Fatal("expected *HTTPServer")
	}

	req := httptest.NewRequest("GET", "/hello", nil)
	resp, err := httpSrv.app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body["msg"] != "world" {
		t.Fatalf("expected msg=world, got %v", body)
	}
}

func TestNewHTTP_ErrorHandler(t *testing.T) {
	var handlerCalled bool
	cfg := HTTPConfig{
		AppName: "test",
		Port:    "8080",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			handlerCalled = true
			return c.Status(fiber.StatusTeapot).SendString("custom error")
		},
	}
	routes := testRoutes{fn: func(app *fiber.App) {
		app.Get("/err", func(c *fiber.Ctx) error {
			return errors.New("something broke")
		})
	}}

	srv := NewHTTP(cfg, routes, testLog())
	httpSrv, ok := srv.(*HTTPServer)
	if !ok {
		t.Fatal("expected *HTTPServer")
	}

	req := httptest.NewRequest("GET", "/err", nil)
	resp, err := httpSrv.app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusTeapot {
		t.Fatalf("expected 418, got %d", resp.StatusCode)
	}
	if !handlerCalled {
		t.Fatal("error handler was not called")
	}
}

func TestHTTPServer_Name(t *testing.T) {
	cfg := HTTPConfig{AppName: "test", Port: "8080"}
	srv := NewHTTP(cfg, testRoutes{fn: func(app *fiber.App) {}}, testLog())
	httpSrv, ok := srv.(*HTTPServer)
	if !ok {
		t.Fatal("expected *HTTPServer")
	}
	if got := httpSrv.Name(); got != "http" {
		t.Fatalf("expected 'http', got %q", got)
	}
}

func TestNewHTTP_AppliesMiddleware(t *testing.T) {
	cfg := HTTPConfig{
		AppName:     "test",
		Port:        "8080",
		CORSOrigins: "https://example.com",
	}
	routes := testRoutes{fn: func(app *fiber.App) {
		app.Get("/cors", func(c *fiber.Ctx) error {
			return c.SendString("ok")
		})
	}}

	srv := NewHTTP(cfg, routes, testLog())
	httpSrv, ok := srv.(*HTTPServer)
	if !ok {
		t.Fatal("expected *HTTPServer")
	}

	req := httptest.NewRequest("GET", "/cors", nil)
	req.Header.Set("Origin", "https://example.com")
	req.Header.Set("Access-Control-Request-Method", "GET")
	resp, err := httpSrv.app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}

	if resp.Header.Get("Access-Control-Allow-Origin") != "https://example.com" {
		t.Fatalf("expected CORS origin header, got %q", resp.Header.Get("Access-Control-Allow-Origin"))
	}
}

func TestNewHTTP_RecoverPanic(t *testing.T) {
	cfg := HTTPConfig{
		AppName: "test",
		Port:    "8080",
	}
	routes := testRoutes{fn: func(app *fiber.App) {
		app.Get("/panic", func(c *fiber.Ctx) error {
			panic("test panic")
		})
	}}

	srv := NewHTTP(cfg, routes, testLog())
	httpSrv, ok := srv.(*HTTPServer)
	if !ok {
		t.Fatal("expected *HTTPServer")
	}

	req := httptest.NewRequest("GET", "/panic", nil)
	resp, err := httpSrv.app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("expected 500 for panic recovery, got %d", resp.StatusCode)
	}
}

func TestNewHTTP_PortEmpty(t *testing.T) {
	srv := NewHTTP(HTTPConfig{}, nil, testLog())
	if _, ok := srv.(noop); !ok {
		t.Fatal("empty port should return noop")
	}
}
