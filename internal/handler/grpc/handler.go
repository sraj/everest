package grpc

import (
	"github.com/sraj/everest/internal/service"

	"google.golang.org/grpc"
)

type Handler struct {
	docService service.DocumentService
}

func New(docService service.DocumentService) *Handler {
	return &Handler{
		docService: docService,
	}
}

func (h *Handler) RegisterServices(s *grpc.Server) {
	// Register your generated protobuf service implementations here.
	// Example:
	//   pb.RegisterDocumentServiceServer(s, &documentServer{docService: h.docService})
	_ = h.docService
}
