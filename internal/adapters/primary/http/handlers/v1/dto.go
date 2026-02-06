package v1

// CreateMetadataRequest is the DTO for the POST /metadata endpoint
type CreateMetadataRequest struct {
	URL string `json:"url" binding:"required,url"`
}

// MetadataResponse is the DTO for the GET /metadata endpoint and responses
type MetadataResponse struct {
	URL        string            `json:"url" example:"https://example.com"`
	Headers    map[string]string `json:"headers"`
	Cookies    map[string]string `json:"cookies"`
	PageSource string            `json:"page_source" example:"<html>...</html>"`
}
