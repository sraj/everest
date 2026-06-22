package server

import (
	"context"
	"log/slog"
	"os"
	"testing"
	"time"
)

type testServer struct {
	name     string
	started  chan struct{}
	shutdown chan struct{}
}

func (s *testServer) Name() string                   { return s.name }
func (s *testServer) Start() error                   { close(s.started); <-s.shutdown; return nil }
func (s *testServer) Shutdown(_ context.Context) error { close(s.shutdown); return nil }

func TestRun_StartAndShutdown(t *testing.T) {
	srv := &testServer{
		name:     "test",
		started:  make(chan struct{}),
		shutdown: make(chan struct{}),
	}

	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))

	done := make(chan error, 1)
	go func() {
		done <- Run(log, srv)
	}()

	select {
	case <-srv.started:
	case <-time.After(2 * time.Second):
		t.Fatal("server did not start")
	}

	if err := srv.Shutdown(context.Background()); err != nil {
		t.Fatalf("shutdown failed: %v", err)
	}

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not return after shutdown")
	}
}

func TestRun_NoServers(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	if err := Run(log); err != nil {
		t.Fatalf("Run with no servers returned error: %v", err)
	}
}

func TestRun_SkipsNoop(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	if err := Run(log, noop{}, noop{}); err != nil {
		t.Fatalf("Run with only noops returned error: %v", err)
	}
}

func TestNewHTTP_Disabled(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	srv := NewHTTP(HTTPConfig{}, nil, log)
	if _, ok := srv.(noop); !ok {
		t.Fatal("NewHTTP with empty port should return noop")
	}
}

func TestNewGRPC_Disabled(t *testing.T) {
	srv := NewGRPC(GRPCConfig{}, nil)
	if _, ok := srv.(noop); !ok {
		t.Fatal("NewGRPC with empty port should return noop")
	}
}
