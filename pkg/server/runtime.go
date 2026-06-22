package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// Server is a transport server (HTTP, gRPC, etc).
type Server interface {
	Name() string
	Start() error
	Shutdown(ctx context.Context) error
}

// noop is a disabled server that does nothing.
type noop struct{}

func (noop) Name() string                     { return "noop" }
func (noop) Start() error                     { return nil }
func (noop) Shutdown(_ context.Context) error { return nil }

// Run starts all provided servers and handles graceful shutdown on OS signals.
// Servers that are nil or noop are silently skipped.
func Run(log *slog.Logger, servers ...Server) error {
	var active []Server
	for _, s := range servers {
		if s == nil {
			continue
		}
		if _, ok := s.(noop); ok {
			continue
		}
		active = append(active, s)
	}

	if len(active) == 0 {
		log.Info("no servers configured")
		return nil
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errCh := make(chan error, len(active))

	for _, srv := range active {
		s := srv
		log.Info("starting server", "server", s.Name())
		go func() {
			if err := s.Start(); err != nil {
				errCh <- fmt.Errorf("%s server failed: %w", s.Name(), err)
			} else {
				errCh <- nil
			}
		}()
	}

	select {
	case <-ctx.Done():
		log.Info("shutdown signal received")
		return shutdownAll(log, active...)
	case err := <-errCh:
		if ctx.Err() != nil {
			return shutdownAll(log, active...)
		}
		if err != nil {
			log.Error("server terminated", "error", err.Error())
			_ = shutdownAll(log, active...)
		}
		return err
	}
}

func shutdownAll(log *slog.Logger, servers ...Server) error {
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var errs []error
	for i := len(servers) - 1; i >= 0; i-- {
		srv := servers[i]
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("failed to shutdown server", "server", srv.Name(), "error", err)
			errs = append(errs, fmt.Errorf("%s shutdown failed: %w", srv.Name(), err))
		}
	}

	if len(errs) == 0 {
		return nil
	}
	return errors.Join(errs...)
}
