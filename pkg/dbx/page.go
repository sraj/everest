package dbx

import "context"

// SortDir is the ORDER BY direction.
type SortDir string

const (
	ASC  SortDir = "ASC"
	DESC SortDir = "DESC"
)

// Page carries pagination parameters (1-based page number).
type Page struct {
	Number int // 1-based; defaults to 1
	Size   int // rows per page; defaults to 20
}

func (p Page) offset() int {
	if p.Number < 1 {
		p.Number = 1
	}
	return (p.Number - 1) * p.size()
}

func (p Page) size() int {
	if p.Size < 1 {
		return 20
	}
	return p.Size
}

// PageResult wraps a typed list response with pagination metadata.
type PageResult[T any] struct {
	Items      []T
	Total      int
	Page       int
	PageSize   int
	TotalPages int
}

// Paginate runs a COUNT(*) query and a data query against the same
// SelectBuilder predicate and returns a fully populated PageResult[T].
//
// T must be a struct whose fields are tagged with `db:` to match column names.
func Paginate[T any](ctx context.Context, b SelectBuilder, p Page) (*PageResult[T], error) {
	total, err := b.Count(ctx)
	if err != nil {
		return nil, err
	}

	var items []T
	if err := b.Paginate(p).All(ctx, &items); err != nil {
		return nil, err
	}

	sz := p.size()
	totalPages := 0
	if sz > 0 {
		totalPages = (total + sz - 1) / sz
	}

	return &PageResult[T]{
		Items:      items,
		Total:      total,
		Page:       p.Number,
		PageSize:   sz,
		TotalPages: totalPages,
	}, nil
}
