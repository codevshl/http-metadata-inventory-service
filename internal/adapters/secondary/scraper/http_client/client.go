package http_client

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/codevshl/http-metadata-inventory-service/internal/core/apperror"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/domain"
	"github.com/codevshl/http-metadata-inventory-service/internal/core/ports"
)

type scraper struct {
	client *http.Client
}

// NewScraper creates a new HTTP scraper with configured timeouts
func NewScraper() ports.Scraper {
	return &scraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return fmt.Errorf("stopped after 10 redirects")
				}
				return nil
			},
		},
	}
}

func (s *scraper) Fetch(ctx context.Context, url string) (*domain.Metadata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrValidation, "invalid url", err)
	}

	req.Header.Set("User-Agent", "MetadataInventoryBot/1.0")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrUpstream, "failed to fetch metadata", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, apperror.Wrap(apperror.ErrUpstream, "upstream returned error", fmt.Errorf("unexpected status code: %d", resp.StatusCode))
	}

	headers := make(map[string]string)
	for key, values := range resp.Header {
		if len(values) > 0 {
			headers[key] = values[0]
		}
	}

	cookies := make(map[string]string)
	for _, cookie := range resp.Cookies() {
		cookies[cookie.Name] = cookie.Value
	}

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, apperror.Wrap(apperror.ErrUpstream, "failed to read upstream response", err)
	}

	return &domain.Metadata{
		URL:        url,
		Headers:    headers,
		Cookies:    cookies,
		PageSource: string(bodyBytes),
	}, nil
}
