package valueobject

import (
	"net/url"
	"strings"

	"github.com/leonardo-gorska/nexuslink/internal/domain"
)

// URL represents a validated URL string
type URL struct {
	Raw string
}

// NewURL validates and normalizes the input URL
func NewURL(rawURL string) (*URL, error) {
	rawURL = strings.TrimSpace(rawURL)

	if rawURL == "" {
		return nil, domain.ErrInvalidURL
	}

	u, err := url.ParseRequestURI(rawURL)
	if err != nil || u.Scheme == "" || u.Host == "" {
		// Try adding https:// if scheme is missing
		if !strings.Contains(rawURL, "://") {
			rawURL = "https://" + rawURL
			u, err = url.ParseRequestURI(rawURL)
			if err != nil || u.Scheme == "" || u.Host == "" {
				return nil, domain.ErrInvalidURL
			}
		} else {
			return nil, domain.ErrInvalidURL
		}
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, domain.ErrInvalidURL
	}

	return &URL{Raw: rawURL}, nil
}
