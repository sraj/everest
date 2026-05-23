package logger

import (
	"context"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog"
)

type zerologHandler struct {
	logger zerolog.Logger
	lvl    slog.Level
}

func (h *zerologHandler) Handle(_ context.Context, r slog.Record) error {
	evt := h.logger.WithLevel(slogToZerologLevel(r.Level))
	r.Attrs(func(a slog.Attr) bool {
		switch a.Value.Kind() {
		case slog.KindString:
			evt = evt.Str(a.Key, a.Value.String())
		case slog.KindInt64:
			evt = evt.Int64(a.Key, a.Value.Int64())
		case slog.KindUint64:
			evt = evt.Uint64(a.Key, a.Value.Uint64())
		case slog.KindFloat64:
			evt = evt.Float64(a.Key, a.Value.Float64())
		case slog.KindBool:
			evt = evt.Bool(a.Key, a.Value.Bool())
		case slog.KindDuration:
			evt = evt.Dur(a.Key, a.Value.Duration())
		case slog.KindTime:
			evt = evt.Time(a.Key, a.Value.Time())
		case slog.KindAny:
			evt = evt.Interface(a.Key, a.Value.Any())
		}
		return true
	})
	evt.Msg(r.Message)
	return nil
}

func (h *zerologHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.lvl
}

func (h *zerologHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	l := h.logger
	for _, a := range attrs {
		l = l.With().Interface(a.Key, a.Value.Any()).Logger()
	}
	return &zerologHandler{logger: l, lvl: h.lvl}
}

func (h *zerologHandler) WithGroup(_ string) slog.Handler {
	return h
}

func slogToZerologLevel(lvl slog.Level) zerolog.Level {
	switch {
	case lvl >= slog.LevelError:
		return zerolog.ErrorLevel
	case lvl >= slog.LevelWarn:
		return zerolog.WarnLevel
	case lvl >= slog.LevelInfo:
		return zerolog.InfoLevel
	default:
		return zerolog.DebugLevel
	}
}

// New creates a new slog logger backed by zerolog's pretty ConsoleWriter
func New(level string, serviceName string) *slog.Logger {
	var lvl slog.Level
	switch strings.ToLower(level) {
	case "debug":
		lvl = slog.LevelDebug
	case "info":
		lvl = slog.LevelInfo
	case "warn":
		lvl = slog.LevelWarn
	case "error":
		lvl = slog.LevelError
	default:
		lvl = slog.LevelInfo
	}

	output := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
		NoColor:    false,
	}

	zlog := zerolog.New(output).
		Level(zerolog.DebugLevel).
		With().
		Timestamp().
		Str("service", serviceName).
		Logger()

	return slog.New(&zerologHandler{logger: zlog, lvl: lvl})
}

// Middleware returns a Fiber middleware that logs requests
func Middleware(log *slog.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := time.Now()
		err := c.Next()
		log.Info("request",
			"method", c.Method(),
			"path", c.Path(),
			"status", c.Response().StatusCode(),
			"ip", c.IP(),
			"request_id", c.GetRespHeader("X-Request-ID"),
			"latency", time.Since(start),
		)
		return err
	}
}
