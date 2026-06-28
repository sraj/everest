package grpc

import (
	"context"
	"errors"
	"log/slog"

	documentsv1 "github.com/sraj/everest/api/gen/go/documents/v1"
	"github.com/sraj/everest/internal/apperror"
	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/service"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Handler registers gRPC services on a gRPC server.
type Handler struct {
	docService service.DocumentService
	log        *slog.Logger
}

func New(docService service.DocumentService, log *slog.Logger) *Handler {
	return &Handler{docService: docService, log: log}
}

func (h *Handler) RegisterServices(s *grpc.Server) {
	svr := &documentServer{docService: h.docService, log: h.log}
	documentsv1.RegisterDocumentServiceServer(s, svr)
}

type documentServer struct {
	documentsv1.UnimplementedDocumentServiceServer
	docService service.DocumentService
	log        *slog.Logger
}

func (s *documentServer) Create(ctx context.Context, req *documentsv1.CreateRequest) (*documentsv1.Document, error) {
	doc, err := s.docService.Create(ctx, service.CreateDocumentInput{
		Title:       req.Title,
		OwnerID:     req.OwnerId,
		Content:     req.Content,
		ContentType: req.ContentType,
	})
	if err != nil {
		s.log.Error("grpc: failed to create document", "error", err.Error())
		return nil, toGRPCError(err)
	}
	return toProtoDocument(doc), nil
}

func (s *documentServer) Get(ctx context.Context, req *documentsv1.GetRequest) (*documentsv1.Document, error) {
	doc, err := s.docService.GetByID(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return toProtoDocument(doc), nil
}

func (s *documentServer) GetContent(ctx context.Context, req *documentsv1.GetRequest) (*documentsv1.GetContentResponse, error) {
	content, err := s.docService.GetContent(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &documentsv1.GetContentResponse{Content: content}, nil
}

func (s *documentServer) List(ctx context.Context, req *documentsv1.ListRequest) (*documentsv1.ListResponse, error) {
	result, err := s.docService.List(ctx, model.Page{Number: int(req.Page), Size: int(req.PageSize)})
	if err != nil {
		return nil, toGRPCError(err)
	}
	items := make([]*documentsv1.Document, 0, len(result.Items))
	for _, doc := range result.Items {
		items = append(items, toProtoDocument(doc))
	}
	return &documentsv1.ListResponse{
		Items:      items,
		Total:      int32(result.Total),
		Page:       int32(result.Page),
		PageSize:   int32(result.PageSize),
		TotalPages: int32(result.TotalPages),
	}, nil
}

func (s *documentServer) Update(ctx context.Context, req *documentsv1.UpdateRequest) (*documentsv1.Document, error) {
	doc, err := s.docService.Update(ctx, service.UpdateDocumentInput{
		ID:      req.Id,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		s.log.Error("grpc: failed to update document", "error", err.Error())
		return nil, toGRPCError(err)
	}
	return toProtoDocument(doc), nil
}

func (s *documentServer) Delete(ctx context.Context, req *documentsv1.DeleteRequest) (*emptypb.Empty, error) {
	if err := s.docService.Delete(ctx, req.Id); err != nil {
		return nil, toGRPCError(err)
	}
	return &emptypb.Empty{}, nil
}

func (s *documentServer) GetThumbnail(ctx context.Context, req *documentsv1.GetRequest) (*documentsv1.GetThumbnailResponse, error) {
	thumbnail, err := s.docService.GetThumbnail(ctx, req.Id)
	if err != nil {
		return nil, toGRPCError(err)
	}
	return &documentsv1.GetThumbnailResponse{Thumbnail: thumbnail}, nil
}

func toProtoDocument(doc *model.Document) *documentsv1.Document {
	d := &documentsv1.Document{
		Id:        doc.ID,
		Title:     doc.Title,
		OwnerId:   doc.OwnerID,
		ContentId: doc.ContentID,
		CreatedAt: timestamppb.New(doc.CreatedAt),
		UpdatedAt: timestamppb.New(doc.UpdatedAt),
	}
	if doc.ThumbnailID != nil {
		d.ThumbnailId = doc.ThumbnailID
	}
	return d
}

func toGRPCError(err error) error {
	var ae *apperror.AppError
	if errors.As(err, &ae) {
		return status.Error(httpToGRPCCode(ae.Status), ae.Message)
	}
	return status.Error(codes.Internal, err.Error())
}

func httpToGRPCCode(httpStatus int) codes.Code {
	switch httpStatus {
	case 400:
		return codes.InvalidArgument
	case 401:
		return codes.Unauthenticated
	case 403:
		return codes.PermissionDenied
	case 404:
		return codes.NotFound
	case 409:
		return codes.AlreadyExists
	case 429:
		return codes.ResourceExhausted
	default:
		return codes.Internal
	}
}
