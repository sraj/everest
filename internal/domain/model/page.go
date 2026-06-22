package model

// Page represents pagination input parameters
type Page struct {
	Number int // 1-based page number
	Size   int // items per page
}

// PageResult wraps a list response with pagination metadata
type PageResult struct {
	Items      []*Document `json:"items"`
	Total      int         `json:"total"`
	Page       int         `json:"page"`
	PageSize   int         `json:"page_size"`
	TotalPages int         `json:"total_pages"`
}

// DefaultPage returns sensible pagination defaults
func DefaultPage() Page {
	return Page{Number: 1, Size: 20}
}

// Offset returns the 0-based offset for this page
func (p Page) Offset() int {
	if p.Number < 1 {
		return 0
	}
	return (p.Number - 1) * p.Size
}
