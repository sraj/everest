package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/chromedp/chromedp"
)

// ThumbnailConfig holds configuration for thumbnail generation
type ThumbnailConfig struct {
	Width   int // Viewport width for rendering
	Height  int // Viewport height for rendering
	Quality int // JPEG quality (1-100)
}

// DefaultThumbnailConfig returns sensible defaults for document thumbnails
// Uses 3:4 portrait aspect ratio to match the frontend preview cards
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
}

// thumbnailService handles document thumbnail generation
type thumbnailService struct {
	config ThumbnailConfig
	log    *slog.Logger
}

// NewThumbnailService creates a new thumbnail service
func NewThumbnailService(config ThumbnailConfig, log *slog.Logger) ThumbnailService {
	return &thumbnailService{
		config: config,
		log:    log,
	}
}

// GenerateFromHTML renders HTML content and captures a screenshot as PNG
func (s *thumbnailService) GenerateFromHTML(ctx context.Context, htmlContent []byte) ([]byte, error) {
	// Create a timeout context for the browser operation
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	// Create chromedp context with headless browser
	allocCtx, allocCancel := chromedp.NewExecAllocator(ctx,
		append(chromedp.DefaultExecAllocatorOptions[:],
			chromedp.Flag("headless", true),
			chromedp.Flag("disable-gpu", true),
			chromedp.Flag("no-sandbox", true),
			chromedp.Flag("disable-dev-shm-usage", true),
		)...,
	)
	defer allocCancel()

	browserCtx, browserCancel := chromedp.NewContext(allocCtx)
	defer browserCancel()

	// Wrap HTML in a complete document with proper styling
	fullHTML := s.wrapHTML(htmlContent)

	var screenshot []byte
	err := chromedp.Run(browserCtx,
		chromedp.EmulateViewport(int64(s.config.Width), int64(s.config.Height)),
		chromedp.Navigate("about:blank"),
		chromedp.ActionFunc(func(ctx context.Context) error {
			return chromedp.Evaluate(fmt.Sprintf(`document.write(%q); document.close();`, fullHTML), nil).Do(ctx)
		}),
		chromedp.Sleep(100*time.Millisecond), // Allow rendering to complete
		chromedp.FullScreenshot(&screenshot, s.config.Quality),
	)
	if err != nil {
		s.log.Error("failed to generate thumbnail", "error", err)
		return nil, fmt.Errorf("thumbnail generation failed: %w", err)
	}

	s.log.Debug("thumbnail generated", "size", len(screenshot))
	return screenshot, nil
}

// wrapHTML ensures the HTML content is a complete document with proper styling
// Styled to look like a Google Docs document page
func (s *thumbnailService) wrapHTML(content []byte) string {
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
        h1 {
            font-size: 24pt;
            font-weight: 400;
            margin-bottom: 12pt;
            color: #000;
        }
        h2 {
            font-size: 18pt;
            font-weight: 400;
            margin-bottom: 10pt;
            color: #000;
        }
        h3 {
            font-size: 14pt;
            font-weight: 700;
            margin-bottom: 8pt;
            color: #000;
        }
        p {
            margin-bottom: 11pt;
        }
        ul, ol {
            margin-left: 36pt;
            margin-bottom: 11pt;
        }
        li {
            margin-bottom: 0;
        }
        img {
            max-width: 100%%;
            height: auto;
        }
        table {
            border-collapse: collapse;
            width: 100%%;
            margin-bottom: 11pt;
        }
        th, td {
            border: 1px solid #000;
            padding: 5pt 10pt;
            text-align: left;
        }
        code {
            font-family: 'Courier New', monospace;
            font-size: 10pt;
            background: #f5f5f5;
            padding: 1pt 3pt;
        }
        pre {
            font-family: 'Courier New', monospace;
            font-size: 10pt;
            background: #f5f5f5;
            padding: 10pt;
            margin-bottom: 11pt;
            overflow: hidden;
            white-space: pre-wrap;
        }
        blockquote {
            border-left: 3px solid #ccc;
            margin: 0 0 11pt 0;
            padding-left: 10pt;
            color: #666;
        }
        a {
            color: #1a73e8;
            text-decoration: underline;
        }
    </style>
</head>
<body>
%s
</body>
</html>`, s.config.Width, s.config.Height, s.config.Width, s.config.Height, s.config.Width, s.config.Height, string(content))
}
