package http

// CreateDocumentRequest is the validated body for POST /api/v1/documents.
type CreateDocumentRequest struct {
	Title   string `json:"title"    validate:"max=500"`
	Content string `json:"content"`
}

// UpdateDocumentRequest is the validated body for PUT /api/v1/documents/:id.
type UpdateDocumentRequest struct {
	Title   *string `json:"title"   validate:"omitempty,max=500"`
	Content *string `json:"content"`
}

// ListDocumentsQuery holds pagination query parameters.
type ListDocumentsQuery struct {
	Page int `query:"page" validate:"omitempty,gte=1"`
	Size int `query:"size" validate:"omitempty,gte=1,lte=100"`
}

// MultipartBody holds the parsed fields from a multipart/form-data upload.
type MultipartBody struct {
	Title       string
	Content     []byte
	ContentType string
	FileName    string
}
