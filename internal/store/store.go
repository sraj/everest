package store

import (
	"context"
)

// Store is the single entry point for all data access.
type Store interface {
	Document() DocumentStore
	Content()  ContentStore
	Vector()   VectorStore
	Close()    error
	Ping(ctx context.Context) error
}

type composite struct {
	doc     DocumentStore
	content ContentStore
	vector  VectorStore
	closer  func() error
	pinger  func(ctx context.Context) error
}

// New returns a Store backed by the given sub-stores.
func New(doc DocumentStore, content ContentStore, vector VectorStore, closer func() error, pinger func(ctx context.Context) error) Store {
	return &composite{
		doc:     doc,
		content: content,
		vector:  vector,
		closer:  closer,
		pinger:  pinger,
	}
}

func (s *composite) Document() DocumentStore                   { return s.doc }
func (s *composite) Content() ContentStore                     { return s.content }
func (s *composite) Vector() VectorStore                       { return s.vector }
func (s *composite) Close() error                              { return s.closer() }
func (s *composite) Ping(ctx context.Context) error            { return s.pinger(ctx) }
