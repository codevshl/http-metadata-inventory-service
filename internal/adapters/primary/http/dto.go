package http

// CreateMetadataRequest is the DTO for the POST /metadata endpoint
type CreateMetadataRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// MetadataResponse is the DTO for the GET /metadata endpoint and responses
type MetadataResponse struct {
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	Cookies    map[string]string `json:"cookies"`
	PageSource string            `json:"page_source"`
}

// ErrorResponse Standard error response format
type ErrorResponse struct {
	Error string `json:"error"`
}
