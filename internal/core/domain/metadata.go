package domain

// Metadata represents the metadata collected for a given URL.
type Metadata struct {
	URL        string            `json:"url"`
	Headers    map[string]string `json:"headers"`
	Cookies    map[string]string `json:"cookies"`
	PageSource string            `json:"page_source"`
}
