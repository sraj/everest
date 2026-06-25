package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sraj/everest/internal/domain/model"
	"github.com/sraj/everest/internal/store"
)

func docSvcLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
}

type mockDocumentStore struct {
	createFn     func(ctx context.Context, doc *model.Document) error
	getByIDFn    func(ctx context.Context, id string) (*model.Document, error)
	updateFn     func(ctx context.Context, doc *model.Document) error
	deleteFn     func(ctx context.Context, id string) error
	listFn       func(ctx context.Context, page model.Page) (*model.PageResult, error)
	listByOwnerFn func(ctx context.Context, ownerID string, page model.Page) (*model.PageResult, error)
	countFn      func(ctx context.Context) (int, error)
}

func (m *mockDocumentStore) Create(ctx context.Context, doc *model.Document) error {
	return m.createFn(ctx, doc)
}

func (m *mockDocumentStore) GetByID(ctx context.Context, id string) (*model.Document, error) {
	return m.getByIDFn(ctx, id)
}

func (m *mockDocumentStore) GetByOwnerID(ctx context.Context, ownerID string) ([]*model.Document, error) {
	return nil, nil
}

func (m *mockDocumentStore) Update(ctx context.Context, doc *model.Document) error {
	return m.updateFn(ctx, doc)
}

func (m *mockDocumentStore) Delete(ctx context.Context, id string) error {
	return m.deleteFn(ctx, id)
}

func (m *mockDocumentStore) List(ctx context.Context, page model.Page) (*model.PageResult, error) {
	return m.listFn(ctx, page)
}

func (m *mockDocumentStore) ListByOwner(ctx context.Context, ownerID string, page model.Page) (*model.PageResult, error) {
	return m.listByOwnerFn(ctx, ownerID, page)
}

func (m *mockDocumentStore) Count(ctx context.Context) (int, error) {
	return m.countFn(ctx)
}

type mockContentStore struct {
	saveFn   func(ctx context.Context, key string, content []byte, contentType string) error
	getFn    func(ctx context.Context, key string) ([]byte, error)
	deleteFn func(ctx context.Context, key string) error
}

func (m *mockContentStore) Save(ctx context.Context, key string, content []byte, contentType string) error {
	return m.saveFn(ctx, key, content, contentType)
}

func (m *mockContentStore) Get(ctx context.Context, key string) ([]byte, error) {
	return m.getFn(ctx, key)
}

func (m *mockContentStore) Delete(ctx context.Context, key string) error {
	return m.deleteFn(ctx, key)
}

func noopClose() error                       { return nil }
func noopPing(_ context.Context) error       { return nil }

const testOwnerID = "user-1"

type mockThumbnailSvc struct {
	mu                sync.Mutex
	generateFn        func(ctx context.Context, html []byte) ([]byte, error)
	closeFn           func()
	thumbnailCaptured chan struct{}
}

func (m *mockThumbnailSvc) GenerateFromHTML(ctx context.Context, html []byte) ([]byte, error) {
	if m.thumbnailCaptured != nil {
		defer func() { m.thumbnailCaptured <- struct{}{} }()
	}
	return m.generateFn(ctx, html)
}

func (m *mockThumbnailSvc) Close() {
	if m.closeFn != nil {
		m.closeFn()
	}
}

func sampleDocument() *model.Document {
	return &model.Document{
		ID:      uuid.New().String(),
		Title:   "Test Doc",
		OwnerID: testOwnerID,
		ContentID: uuid.New().String(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestDocumentService_Create(t *testing.T) {
	contentStore := &mockContentStore{
		saveFn: func(ctx context.Context, key string, content []byte, contentType string) error {
			assert.Equal(t, "text/html", contentType)
			return nil
		},
	}
	docStore := &mockDocumentStore{
		createFn: func(ctx context.Context, doc *model.Document) error {
			assert.NotEmpty(t, doc.ID)
			assert.Equal(t, "My Doc", doc.Title)
			return nil
		},
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	doc, err := svc.Create(context.Background(), CreateDocumentInput{
		Title:   "My Doc",
		OwnerID: "user-1",
		Content: []byte("<p>hello</p>"),
	}, testOwnerID)
	require.NoError(t, err)
	assert.Equal(t, "My Doc", doc.Title)
	assert.NotEmpty(t, doc.ID)
	assert.NotEmpty(t, doc.ContentID)
}

func TestDocumentService_Create_ContentTypeDefaults(t *testing.T) {
	var capturedType string
	contentStore := &mockContentStore{
		saveFn: func(ctx context.Context, key string, content []byte, contentType string) error {
			capturedType = contentType
			return nil
		},
	}
	docStore := &mockDocumentStore{
		createFn: func(ctx context.Context, doc *model.Document) error { return nil },
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	_, err := svc.Create(context.Background(), CreateDocumentInput{
		Title:   "Doc",
		OwnerID: "user-1",
		Content: []byte("<p>hi</p>"),
	}, testOwnerID)
	require.NoError(t, err)
	assert.Equal(t, "text/html", capturedType)
}

func TestDocumentService_Create_ContentSaveFails(t *testing.T) {
	contentStore := &mockContentStore{
		saveFn: func(ctx context.Context, key string, content []byte, contentType string) error {
			return errors.New("minio down")
		},
	}
	docStore := &mockDocumentStore{}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	_, err := svc.Create(context.Background(), CreateDocumentInput{
		Title:   "Doc",
		OwnerID: "user-1",
		Content: []byte("<p>hi</p>"),
	}, testOwnerID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "minio down")
}

func TestDocumentService_Create_DocCreateFails_ContentCleanup(t *testing.T) {
	var contentDeleted string
	contentStore := &mockContentStore{
		saveFn:   func(ctx context.Context, key string, content []byte, contentType string) error { return nil },
		deleteFn: func(ctx context.Context, key string) error { contentDeleted = key; return nil },
	}
	docStore := &mockDocumentStore{
		createFn: func(ctx context.Context, doc *model.Document) error {
			return errors.New("db constraint")
		},
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	_, err := svc.Create(context.Background(), CreateDocumentInput{
		Title:   "Doc",
		OwnerID: "user-1",
		Content: []byte("<p>hi</p>"),
	}, testOwnerID)
	require.Error(t, err)
	assert.NotEmpty(t, contentDeleted, "content should be cleaned up on doc creation failure")
}

func TestDocumentService_Create_WithThumbnail(t *testing.T) {
	thumbnailCaptured := make(chan struct{}, 1)
	thumbSvc := &mockThumbnailSvc{
		thumbnailCaptured: thumbnailCaptured,
		generateFn: func(ctx context.Context, html []byte) ([]byte, error) {
			return []byte("png-data"), nil
		},
	}

	contentStore := &mockContentStore{
		saveFn: func(ctx context.Context, key string, content []byte, contentType string) error {
			return nil
		},
	}
	docStore := &mockDocumentStore{
		createFn: func(ctx context.Context, doc *model.Document) error { return nil },
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) {
			return &model.Document{
				ID:        id,
				Title:     "Doc",
				OwnerID:   "user-1",
				ContentID: uuid.New().String(),
				CreatedAt: time.Now(),
				UpdatedAt: time.Now(),
			}, nil
		},
		updateFn: func(ctx context.Context, doc *model.Document) error { return nil },
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), thumbSvc, docSvcLogger())

	_, err := svc.Create(context.Background(), CreateDocumentInput{
		Title:   "Doc",
		OwnerID: "user-1",
		Content: []byte("<p>hi</p>"),
	}, testOwnerID)
	require.NoError(t, err)

	select {
	case <-thumbnailCaptured:
		// async thumbnail was generated
	case <-time.After(2 * time.Second):
		t.Fatal("thumbnail was not generated within timeout")
	}
}

func TestDocumentService_GetByID(t *testing.T) {
	expected := sampleDocument()
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) {
			return expected, nil
		},
	}
	svc := NewDocumentService(store.New(docStore, nil, noopClose, noopPing), nil, docSvcLogger())

	doc, err := svc.GetByID(context.Background(), expected.ID, testOwnerID)
	require.NoError(t, err)
	assert.Equal(t, expected.ID, doc.ID)
}

func TestDocumentService_GetByID_NotFound(t *testing.T) {
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewDocumentService(store.New(docStore, nil, noopClose, noopPing), nil, docSvcLogger())

	_, err := svc.GetByID(context.Background(), "missing", testOwnerID)
	require.Error(t, err)
}

func TestDocumentService_GetContent(t *testing.T) {
	doc := sampleDocument()
	contentStore := &mockContentStore{
		getFn: func(ctx context.Context, key string) ([]byte, error) {
			assert.Equal(t, doc.ContentID, key)
			return []byte("<p>content</p>"), nil
		},
	}
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) {
			return doc, nil
		},
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	data, err := svc.GetContent(context.Background(), doc.ID, testOwnerID)
	require.NoError(t, err)
	assert.Equal(t, []byte("<p>content</p>"), data)
}

func TestDocumentService_GetContent_DocNotFound(t *testing.T) {
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewDocumentService(store.New(docStore, nil, noopClose, noopPing), nil, docSvcLogger())

	_, err := svc.GetContent(context.Background(), "missing", testOwnerID)
	require.Error(t, err)
}

func TestDocumentService_Update(t *testing.T) {
	doc := sampleDocument()
	var updatedDoc *model.Document
	contentStore := &mockContentStore{
		saveFn: func(ctx context.Context, key string, content []byte, contentType string) error {
			return nil
		},
	}
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) {
			return doc, nil
		},
		updateFn: func(ctx context.Context, d *model.Document) error {
			updatedDoc = d
			return nil
		},
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	result, err := svc.Update(context.Background(), UpdateDocumentInput{
		ID:      doc.ID,
		Title:   "Updated Title",
		Content: []byte("<p>new</p>"),
	}, testOwnerID)
	require.NoError(t, err)
	assert.Equal(t, "Updated Title", result.Title)
	assert.Equal(t, "Updated Title", updatedDoc.Title)
}

func TestDocumentService_Update_TitleOnly(t *testing.T) {
	doc := sampleDocument()
	contentStore := &mockContentStore{}
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) { return doc, nil },
		updateFn: func(ctx context.Context, d *model.Document) error { return nil },
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	result, err := svc.Update(context.Background(), UpdateDocumentInput{
		ID:    doc.ID,
		Title: "New Title",
	}, testOwnerID)
	require.NoError(t, err)
	assert.Equal(t, "New Title", result.Title)
}

func TestDocumentService_Update_NotFound(t *testing.T) {
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewDocumentService(store.New(docStore, nil, noopClose, noopPing), nil, docSvcLogger())

	_, err := svc.Update(context.Background(), UpdateDocumentInput{
		ID:    "missing",
		Title: "X",
	}, testOwnerID)
	require.Error(t, err)
}

func TestDocumentService_Update_ContentSaveFails(t *testing.T) {
	doc := sampleDocument()
	contentStore := &mockContentStore{
		saveFn: func(ctx context.Context, key string, content []byte, contentType string) error {
			return errors.New("minio write error")
		},
	}
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) { return doc, nil },
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	_, err := svc.Update(context.Background(), UpdateDocumentInput{
		ID:      doc.ID,
		Content: []byte("<p>new</p>"),
	}, testOwnerID)
	require.Error(t, err)
}

func TestDocumentService_Delete(t *testing.T) {
	doc := sampleDocument()
	var deletedDocID, deletedContentID string
	contentStore := &mockContentStore{
		deleteFn: func(ctx context.Context, key string) error {
			deletedContentID = key
			return nil
		},
	}
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) { return doc, nil },
		deleteFn: func(ctx context.Context, id string) error {
			deletedDocID = id
			return nil
		},
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	err := svc.Delete(context.Background(), doc.ID, testOwnerID)
	require.NoError(t, err)
	assert.Equal(t, doc.ID, deletedDocID)
	assert.Equal(t, doc.ContentID, deletedContentID)
}

func TestDocumentService_Delete_WithThumbnail(t *testing.T) {
	tid := "thumb-123"
	doc := sampleDocument()
	doc.ThumbnailID = &tid

	var deletedKeys []string
	contentStore := &mockContentStore{
		deleteFn: func(ctx context.Context, key string) error {
			deletedKeys = append(deletedKeys, key)
			return nil
		},
	}
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) { return doc, nil },
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	err := svc.Delete(context.Background(), doc.ID, testOwnerID)
	require.NoError(t, err)
	assert.Contains(t, deletedKeys, doc.ContentID)
	assert.Contains(t, deletedKeys, "thumb-123")
}

func TestDocumentService_Delete_ContentDeleteFails(t *testing.T) {
	doc := sampleDocument()
	contentStore := &mockContentStore{
		deleteFn: func(ctx context.Context, key string) error {
			return errors.New("minio error")
		},
	}
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) { return doc, nil },
		deleteFn: func(ctx context.Context, id string) error { return nil },
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	err := svc.Delete(context.Background(), doc.ID, testOwnerID)
	require.NoError(t, err)
}

func TestDocumentService_Delete_NotFound(t *testing.T) {
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) {
			return nil, errors.New("not found")
		},
	}
	svc := NewDocumentService(store.New(docStore, nil, noopClose, noopPing), nil, docSvcLogger())

	err := svc.Delete(context.Background(), "missing", testOwnerID)
	require.Error(t, err)
}

func TestDocumentService_List(t *testing.T) {
	docs := []*model.Document{sampleDocument(), sampleDocument()}
	docStore := &mockDocumentStore{
		listByOwnerFn: func(ctx context.Context, ownerID string, page model.Page) (*model.PageResult, error) {
			return &model.PageResult{
				Items:      docs,
				Total:      5,
				Page:       page.Number,
				PageSize:   page.Size,
				TotalPages: 3,
			}, nil
		},
	}
	svc := NewDocumentService(store.New(docStore, nil, noopClose, noopPing), nil, docSvcLogger())

	result, err := svc.List(context.Background(), model.Page{Number: 1, Size: 2}, testOwnerID)
	require.NoError(t, err)
	assert.Equal(t, 2, len(result.Items))
	assert.Equal(t, 5, result.Total)
	assert.Equal(t, 1, result.Page)
	assert.Equal(t, 2, result.PageSize)
}

func TestDocumentService_GetThumbnail(t *testing.T) {
	tid := "thumb-abc"
	doc := sampleDocument()
	doc.ThumbnailID = &tid

	contentStore := &mockContentStore{
		getFn: func(ctx context.Context, key string) ([]byte, error) {
			assert.Equal(t, "thumb-abc", key)
			return []byte("png-data"), nil
		},
	}
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) { return doc, nil },
	}
	svc := NewDocumentService(store.New(docStore, contentStore, noopClose, noopPing), nil, docSvcLogger())

	data, err := svc.GetThumbnail(context.Background(), doc.ID, testOwnerID)
	require.NoError(t, err)
	assert.Equal(t, []byte("png-data"), data)
}

func TestDocumentService_GetThumbnail_Nil(t *testing.T) {
	doc := sampleDocument() // ThumbnailID is nil

	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) { return doc, nil },
	}
	svc := NewDocumentService(store.New(docStore, nil, noopClose, noopPing), nil, docSvcLogger())

	data, err := svc.GetThumbnail(context.Background(), doc.ID, testOwnerID)
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestDocumentService_GetThumbnail_Empty(t *testing.T) {
	emptyID := ""
	doc := sampleDocument()
	doc.ThumbnailID = &emptyID

	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) { return doc, nil },
	}
	svc := NewDocumentService(store.New(docStore, nil, noopClose, noopPing), nil, docSvcLogger())

	data, err := svc.GetThumbnail(context.Background(), doc.ID, testOwnerID)
	require.NoError(t, err)
	assert.Nil(t, data)
}

func TestDocumentService_GetThumbnail_DocNotFound(t *testing.T) {
	docStore := &mockDocumentStore{
		getByIDFn: func(ctx context.Context, id string) (*model.Document, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	svc := NewDocumentService(store.New(docStore, nil, noopClose, noopPing), nil, docSvcLogger())

	_, err := svc.GetThumbnail(context.Background(), "missing", testOwnerID)
	require.Error(t, err)
}
