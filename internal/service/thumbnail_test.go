package service

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestThumbnailConfig_Defaults(t *testing.T) {
	cfg := DefaultThumbnailConfig()

	assert.Equal(t, 600, cfg.Width)
	assert.Equal(t, 800, cfg.Height)
	assert.Equal(t, 80, cfg.Quality)
	assert.Equal(t, 5, cfg.Workers)
	assert.Equal(t, 20, cfg.Buffer)
}

func TestThumbnailService_Close(t *testing.T) {
	cfg := ThumbnailConfig{Workers: 1, Buffer: 1}
	svc := NewThumbnailService(cfg, testLogger())

	require.NotNil(t, svc)

	svc.Close()

	_, err := svc.GenerateFromHTML(context.Background(), []byte("hi"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "closed")
}

func TestThumbnailService_DoubleClose(t *testing.T) {
	cfg := ThumbnailConfig{Workers: 1, Buffer: 1}
	svc := NewThumbnailService(cfg, testLogger())
	svc.Close()
	svc.Close() // should not panic
}

func TestThumbnailService_ContextCancellation(t *testing.T) {
	cfg := ThumbnailConfig{Workers: 1, Buffer: 1}
	svc := NewThumbnailService(cfg, testLogger())
	defer svc.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := svc.GenerateFromHTML(ctx, []byte("hi"))
	require.Error(t, err)
}

func TestThumbnailService_GenerateFromHTML(t *testing.T) {
	t.Skip("requires headless Chrome — run locally with Chrome installed")

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
	require.NoError(t, err)
	assert.NotEmpty(t, data)
}

func TestThumbnailService_WorkerPool(t *testing.T) {
	t.Skip("requires headless Chrome — run locally with Chrome installed")

	cfg := ThumbnailConfig{
		Width:   100,
		Height:  100,
		Quality: 30,
		Workers: 3,
		Buffer:  10,
	}

	svc := NewThumbnailService(cfg, testLogger())
	defer svc.Close()

	type result struct {
		data []byte
		err  error
	}

	results := make(chan result, 6)
	for i := 0; i < 6; i++ {
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			data, err := svc.GenerateFromHTML(ctx, []byte("<p>worker</p>"))
			results <- result{data, err}
		}()
	}

	for i := 0; i < 6; i++ {
		r := <-results
		require.NoError(t, r.err)
		assert.NotEmpty(t, r.data)
	}
}

func TestThumbnailService_ConcurrentClose(t *testing.T) {
	cfg := ThumbnailConfig{Workers: 2, Buffer: 5}
	svc := NewThumbnailService(cfg, testLogger())

	var done = make(chan struct{})

	go func() {
		svc.Close()
		close(done)
	}()

	go func() {
		svc.Close()
	}()

	<-done
}
