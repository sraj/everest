package logger

import (
	"log/slog"
	"testing"
)

func TestNew(t *testing.T) {
	l := New("info", "test")
	if l == nil {
		t.Fatal("New() returned nil")
	}
	l.Info("test message")
}

func TestNew_Levels(t *testing.T) {
	for _, level := range []string{"debug", "info", "warn", "error", "invalid"} {
		l := New(level, "test")
		l.Info(level + " test")
	}
}

func TestNewJSON(t *testing.T) {
	l := NewJSON("debug", "test")
	if l == nil {
		t.Fatal("NewJSON() returned nil")
	}
}

func TestNew_Logging(t *testing.T) {
	l := New("debug", "test-service")

	l.Debug("debug msg", "key", "val")
	l.Info("info msg", "count", 42)
	l.Warn("warn msg", "flag", true)
	l.Error("error msg", "err", "something failed")

	l.With("component", "test").Info("with attr")

	l.LogAttrs(nil, slog.LevelInfo, "log attrs", slog.String("k", "v"))
}
