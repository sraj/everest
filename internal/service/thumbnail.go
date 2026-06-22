package service

import (
	"context"
	"fmt"
	"log/slog"
	"net/url"
	"sync"
	"time"

	"github.com/chromedp/chromedp"
)

// ThumbnailConfig holds configuration for thumbnail generation
type ThumbnailConfig struct {
	Width   int
	Height  int
	Quality int
}

// DefaultThumbnailConfig returns sensible defaults for document thumbnails
func DefaultThumbnailConfig() ThumbnailConfig {
	return ThumbnailConfig{
		Width:   600,
		Height:  800,
		Quality: 80,
	}
}

// ThumbnailService defines the interface for thumbnail generation
type ThumbnailService interface {
	GenerateFromHTML(ctx context.Context, htmlContent []byte) ([]byte, error)
	Close()
}

// thumbnailService handles document thumbnail generation using a persistent browser.
type thumbnailService struct {
	config      ThumbnailConfig
	log         *slog.Logger
	allocCtx    context.Context
	allocCancel context.CancelFunc
	sem         chan struct{}
	mu          sync.Mutex
	closed      bool
}

// NewThumbnailService creates a new thumbnail service with a long-lived headless browser.
func NewThumbnailService(config ThumbnailConfig, log *slog.Logger) ThumbnailService {
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)...,
	)

	log.Info("thumbnail browser started")
	return &thumbnailService{
		config:      config,
		log:         log,
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		sem:         make(chan struct{}, 2),
	}
}

// GenerateFromHTML renders HTML content and captures a screenshot.
func (s *thumbnailService) GenerateFromHTML(ctx context.Context, htmlContent []byte) ([]byte, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, fmt.Errorf("thumbnail service is closed")
	}
	s.mu.Unlock()

	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	fullHTML := wrapHTML(htmlContent, s.config.Width, s.config.Height)
	dataURL := "data:text/html;charset=utf-8," + url.PathEscape(fullHTML)

	browserCtx, browserCancel := chromedp.NewContext(s.allocCtx)
	defer browserCancel()

	browserCtx, cancel := context.WithTimeout(browserCtx, 30*time.Second)
	defer cancel()

	var screenshot []byte
	err := chromedp.Run(browserCtx,
		chromedp.EmulateViewport(int64(s.config.Width), int64(s.config.Height)),
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body"),
		chromedp.FullScreenshot(&screenshot, s.config.Quality),
	)
	if err != nil {
		s.log.Error("thumbnail generation failed", "error", err)
		return nil, fmt.Errorf("thumbnail generation failed: %w", err)
	}

	s.log.Debug("thumbnail generated", "size", len(screenshot))
	return screenshot, nil
}

func (s *thumbnailService) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return
	}
	s.closed = true
	s.allocCancel()
	s.log.Info("thumbnail browser stopped")
}

// wrapHTML wraps content in a styled document matching Google Docs appearance.
func wrapHTML(content []byte, width, height int) string {
	return fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=%d, height=%d">
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        html {
            width: %dpx;
            height: %dpx;
            background: #f8f9fa;
        }
        body {
            width: %dpx;
            min-height: %dpx;
            margin: 0 auto;
            background: white;
            font-family: 'Arial', sans-serif;
            font-size: 11pt;
            line-height: 1.5;
            color: #000;
            padding: 72px 72px 72px 72px;
            overflow: hidden;
        }
        h1 { font-size: 24pt; font-weight: 400; margin-bottom: 12pt; color: #000; }
        h2 { font-size: 18pt; font-weight: 400; margin-bottom: 10pt; color: #000; }
        h3 { font-size: 14pt; font-weight: 700; margin-bottom: 8pt; color: #000; }
        p { margin-bottom: 11pt; }
        ul, ol { margin-left: 36pt; margin-bottom: 11pt; }
        li { margin-bottom: 0; }
        img { max-width: 100%%; height: auto; }
        table { border-collapse: collapse; width: 100%%; margin-bottom: 11pt; }
        th, td { border: 1px solid #000; padding: 5pt 10pt; text-align: left; }
        code { font-family: 'Courier New', monospace; font-size: 10pt; background: #f5f5f5; padding: 1pt 3pt; }
        pre { font-family: 'Courier New', monospace; font-size: 10pt; background: #f5f5f5; padding: 10pt; margin-bottom: 11pt; overflow: hidden; white-space: pre-wrap; }
        blockquote { border-left: 3px solid #ccc; margin: 0 0 11pt 0; padding-left: 10pt; color: #666; }
        a { color: #1a73e8; text-decoration: underline; }
    </style>
</head>
<body>
%s
</body>
</html>`, width, height, width, height, width, height, string(content))
}
