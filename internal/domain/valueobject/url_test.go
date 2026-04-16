package valueobject_test

import (
	"testing"

	"github.com/leonardo-gorska/nexuslink/internal/domain"
	"github.com/leonardo-gorska/nexuslink/internal/domain/valueobject"
)

func TestNewURL(t *testing.T) {
	tests := []struct {
		name    string
		rawURL  string
		wantURL string
		wantErr error
	}{
		{
			name:    "valid full https URL",
			rawURL:  "https://github.com/leonardo-gorska/nexuslink",
			wantURL: "https://github.com/leonardo-gorska/nexuslink",
			wantErr: nil,
		},
		{
			name:    "valid full http URL",
			rawURL:  "http://example.com/test?query=1",
			wantURL: "http://example.com/test?query=1",
			wantErr: nil,
		},
		{
			name:    "missing scheme fallback to https",
			rawURL:  "example.com/path",
			wantURL: "https://example.com/path",
			wantErr: nil,
		},
		{
			name:    "bare domain fallback to https",
			rawURL:  "google.com",
			wantURL: "https://google.com",
			wantErr: nil,
		},
		{
			name:    "empty string",
			rawURL:  "   ",
			wantErr: domain.ErrInvalidURL,
		},
		{
			name:    "invalid scheme",
			rawURL:  "ftp://example.com",
			wantErr: domain.ErrInvalidURL,
		},
		{
			name:    "invalid url format",
			rawURL:  "https://",
			wantErr: domain.ErrInvalidURL,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := valueobject.NewURL(tt.rawURL)
			if err != tt.wantErr {
				t.Errorf("NewURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if err == nil && got.Raw != tt.wantURL {
				t.Errorf("NewURL() got.Raw = %v, want %v", got.Raw, tt.wantURL)
			}
		})
	}
}
