package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
)

type testServer struct {
	name     string
	startErr error
	once     sync.Once
	shutdown chan struct{}
}

func (s *testServer) Name() string { return s.name }

func (s *testServer) Start() error {
	if s.startErr != nil {
		return s.startErr
	}
	<-s.shutdown
	return nil
}

func (s *testServer) Shutdown(_ context.Context) error {
	s.once.Do(func() { close(s.shutdown) })
	return nil
}

type failingServer struct{}

func (failingServer) Name() string                     { return "fail" }
func (failingServer) Start() error                     { return fmt.Errorf("boom") }
func (failingServer) Shutdown(_ context.Context) error { return nil }

func testLog() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

func TestRun_NoServers(t *testing.T) {
	if err := Run(testLog()); err != nil {
		t.Fatalf("Run with no servers: %v", err)
	}
}

func TestRun_SkipsNoop(t *testing.T) {
	if err := Run(testLog(), noop{}, noop{}); err != nil {
		t.Fatalf("Run with only noops: %v", err)
	}
}

func TestNewHTTP_Disabled(t *testing.T) {
	srv := NewHTTP(HTTPConfig{}, nil, testLog())
	if _, ok := srv.(noop); !ok {
		t.Fatal("empty port should return noop")
	}
}

func TestNewGRPC_Disabled(t *testing.T) {
	srv := NewGRPC(GRPCConfig{}, nil)
	if _, ok := srv.(noop); !ok {
		t.Fatal("empty port should return noop")
	}
}

func TestServer_Shutdown(t *testing.T) {
	srv := &testServer{
		name:     "test",
		shutdown: make(chan struct{}),
	}

	done := make(chan error, 1)
	go func() { done <- srv.Start() }()

	// Shutdown should allow Start to return nil
	srv.Shutdown(context.Background())

	if err := <-done; err != nil {
		t.Fatalf("Start returned error: %v", err)
	}
}

func TestServer_Failure(t *testing.T) {
	srv := failingServer{}
	if err := srv.Start(); err == nil {
		t.Fatal("expected error from failing server")
	}
}

func TestServer_MultipleShutdown(t *testing.T) {
	srv := &testServer{
		name:     "test",
		shutdown: make(chan struct{}),
	}

	go func() { srv.Start() }()

	srv.Shutdown(context.Background())
	srv.Shutdown(context.Background()) // should not panic
}

func TestServer_MixedNoops(t *testing.T) {
	srv := &testServer{
		name:     "real",
		shutdown: make(chan struct{}),
	}

	go func() { srv.Start() }()

	servers := []Server{noop{}, srv, noop{}}
	for _, s := range servers {
		if _, isNoop := s.(noop); isNoop {
			continue
		}
		s.Shutdown(context.Background())
	}
}
