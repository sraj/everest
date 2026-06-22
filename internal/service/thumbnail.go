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
	Workers int
	Buffer  int
}

// DefaultThumbnailConfig returns sensible defaults for document thumbnails
func DefaultThumbnailConfig() ThumbnailConfig {
	return ThumbnailConfig{
		Width:   600,
		Height:  800,
		Quality: 80,
		Workers: 5,
		Buffer:  20,
	}
}

// ThumbnailService defines the interface for thumbnail generation
type ThumbnailService interface {
	GenerateFromHTML(ctx context.Context, htmlContent []byte) ([]byte, error)
	Close()
}

type thumbnailJob struct {
	html   []byte
	result chan thumbnailResult
}

type thumbnailResult struct {
	data []byte
	err  error
}

// thumbnailService handles document thumbnail generation using worker pools.
type thumbnailService struct {
	config      ThumbnailConfig
	log         *slog.Logger
	allocCtx    context.Context
	allocCancel context.CancelFunc
	jobs        chan thumbnailJob
	workers     int
	wg          sync.WaitGroup
	mu          sync.Mutex
	closed      bool
}

// NewThumbnailService creates a new thumbnail service with worker pool.
func NewThumbnailService(config ThumbnailConfig, log *slog.Logger) ThumbnailService {
	allocCtx, allocCancel := chromedp.NewExecAllocator(context.Background(),
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)...,
	)

	s := &thumbnailService{
		config:      config,
		log:         log,
		allocCtx:    allocCtx,
		allocCancel: allocCancel,
		jobs:        make(chan thumbnailJob, config.Buffer),
		workers:     config.Workers,
	}

	s.wg.Add(config.Workers)
	for i := 0; i < config.Workers; i++ {
		go s.worker()
	}

	log.Info("thumbnail service started", "workers", config.Workers, "buffer", config.Buffer)
	return s
}

func (s *thumbnailService) worker() {
	defer s.wg.Done()

	for job := range s.jobs {
		taskCtx, cancel := context.WithTimeout(s.allocCtx, 30*time.Second)
		data, err := s.render(taskCtx, job.html)
		cancel()
		job.result <- thumbnailResult{data: data, err: err}
	}
}

func (s *thumbnailService) render(ctx context.Context, htmlContent []byte) ([]byte, error) {
	fullHTML := wrapHTML(htmlContent, s.config.Width, s.config.Height)
	dataURL := "data:text/html;charset=utf-8," + url.PathEscape(fullHTML)

	browserCtx, browserCancel := chromedp.NewContext(ctx)
	defer browserCancel()

	var screenshot []byte
	err := chromedp.Run(browserCtx,
		chromedp.EmulateViewport(int64(s.config.Width), int64(s.config.Height)),
		chromedp.Navigate(dataURL),
		chromedp.WaitReady("body"),
		chromedp.FullScreenshot(&screenshot, s.config.Quality),
	)
	if err != nil {
		return nil, fmt.Errorf("render failed: %w", err)
	}

	return screenshot, nil
}

// GenerateFromHTML renders HTML content and captures a screenshot.
func (s *thumbnailService) GenerateFromHTML(ctx context.Context, htmlContent []byte) ([]byte, error) {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return nil, fmt.Errorf("thumbnail service is closed")
	}
	s.mu.Unlock()

	job := thumbnailJob{
		html:   htmlContent,
		result: make(chan thumbnailResult, 1),
	}

	select {
	case s.jobs <- job:
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	select {
	case r := <-job.result:
		if r.err != nil {
			s.log.Error("thumbnail generation failed", "error", r.err)
			return nil, r.err
		}
		s.log.Debug("thumbnail generated", "size", len(r.data))
		return r.data, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (s *thumbnailService) Close() {
	s.mu.Lock()
	if s.closed {
		s.mu.Unlock()
		return
	}
	s.closed = true
	s.mu.Unlock()

	close(s.jobs)
	s.wg.Wait()
	s.allocCancel()
	s.log.Info("thumbnail service stopped")
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
