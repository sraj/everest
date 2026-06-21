package server

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
)

// GRPCConfig contains gRPC server configuration.
type GRPCConfig struct {
	Port string
}

// GRPCServiceRegistrar registers gRPC services.
type GRPCServiceRegistrar interface {
	RegisterServices(server *grpc.Server)
}

// GRPCServer wraps gRPC setup and lifecycle.
type GRPCServer struct {
	cfg      GRPCConfig
	listener net.Listener
	server   *grpc.Server
}

// NewGRPC creates a new gRPC server. Returns nil if cfg.Port is empty.
func NewGRPC(cfg GRPCConfig, registrar GRPCServiceRegistrar) Server {
	if cfg.Port == "" {
		return noop{}
	}

	s := grpc.NewServer()
	registrar.RegisterServices(s)

	return &GRPCServer{
		cfg:    cfg,
		server: s,
	}
}

// Name returns the server name.
func (s *GRPCServer) Name() string {
	return "grpc"
}

// Start starts listening for incoming gRPC requests.
func (s *GRPCServer) Start() error {
	listener, err := net.Listen("tcp", ":"+s.cfg.Port)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}
	s.listener = listener

	if err := s.server.Serve(listener); err != nil {
		return fmt.Errorf("serve: %w", err)
	}

	return nil
}

// Shutdown gracefully shuts down the gRPC server.
func (s *GRPCServer) Shutdown(ctx context.Context) error {
	if s.server == nil {
		return nil
	}

	done := make(chan struct{})
	go func() {
		s.server.GracefulStop()
		close(done)
	}()

	select {
	case <-done:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		<-done
		return nil
	}
}
