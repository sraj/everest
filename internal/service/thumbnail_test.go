package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestThumbnailService_GenerateFromHTML(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("requires headless Chrome, skip in CI")
	}

	cfg := ThumbnailConfig{
		Width:   200,
		Height:  200,
		Quality: 50,
		Workers: 1,
		Buffer:  1,
	}

	svc := NewThumbnailService(cfg, testLogger())
	defer svc.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	data, err := svc.GenerateFromHTML(ctx, []byte("<h1>Test</h1>"))
	if err != nil {
		t.Fatalf("GenerateFromHTML failed: %v", err)
	}
	if len(data) == 0 {
		t.Fatal("expected non-empty PNG data")
	}
}

func TestThumbnailService_Closed(t *testing.T) {
	cfg := ThumbnailConfig{Workers: 1, Buffer: 1}
	svc := NewThumbnailService(cfg, testLogger()).(*thumbnailService)
	svc.Close()

	_, err := svc.GenerateFromHTML(context.Background(), []byte("hi"))
	if err == nil {
		t.Fatal("expected error after close")
	}
}

func TestThumbnailConfig_Defaults(t *testing.T) {
	cfg := DefaultThumbnailConfig()

	if cfg.Width != 600 {
		t.Errorf("Width = %d, want 600", cfg.Width)
	}
	if cfg.Height != 800 {
		t.Errorf("Height = %d, want 800", cfg.Height)
	}
	if cfg.Quality != 80 {
		t.Errorf("Quality = %d, want 80", cfg.Quality)
	}
	if cfg.Workers != 5 {
		t.Errorf("Workers = %d, want 5", cfg.Workers)
	}
	if cfg.Buffer != 20 {
		t.Errorf("Buffer = %d, want 20", cfg.Buffer)
	}
}

func TestThumbnailService_WorkerPool(t *testing.T) {
	if os.Getenv("CI") != "" {
		t.Skip("requires headless Chrome, skip in CI")
	}

	cfg := ThumbnailConfig{
		Width:   100,
		Height:  100,
		Quality: 30,
		Workers: 3,
		Buffer:  10,
	}

	svc := NewThumbnailService(cfg, testLogger())
	defer svc.Close()

	var results [6][]byte
	errs := make(chan error, 6)

	for i := 0; i < 6; i++ {
		go func(i int) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			data, err := svc.GenerateFromHTML(ctx, []byte("<p>worker</p>"))
			if err != nil {
				errs <- err
				return
			}
			results[i] = data
			errs <- nil
		}(i)
	}

	for i := 0; i < 6; i++ {
		if err := <-errs; err != nil {
			t.Fatalf("worker %d failed: %v", i, err)
		}
	}

	for i, data := range results {
		if len(data) == 0 {
			t.Errorf("worker %d returned empty data", i)
		}
	}
}
