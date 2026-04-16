package valueobject_test

import (
	"testing"

	"github.com/leonardo-gorska/nexuslink/internal/domain/valueobject"
)

func TestNewHash(t *testing.T) {
	tests := []struct {
		name    string
		length  int
		wantErr bool
	}{
		{
			name:    "length 7",
			length:  7,
			wantErr: false,
		},
		{
			name:    "length 10",
			length:  10,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := valueobject.NewHash(tt.length)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewHash() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(got) != tt.length {
				t.Errorf("NewHash() generated unexpected length: got %d, want %d", len(got), tt.length)
			}
			if !valueobject.ValidateHash(got) {
				// ValidateHash returns false if length is outside 5-10
				if tt.length >= 5 && tt.length <= 10 {
					t.Errorf("NewHash() returned invalid hash format: %s", got)
				}
			}
		})
	}
}

func TestValidateHash(t *testing.T) {
	tests := []struct {
		name string
		hash string
		want bool
	}{
		{"valid hash 7 chars", "aB3xK9z", true},
		{"valid hash 10 chars", "0123456789", true},
		{"valid hash 5 chars", "abcDE", true},
		{"invalid length too short", "abcd", false},
		{"invalid length too long", "abcdefghijk", false},
		{"invalid chars special", "aB3x-9z", false},
		{"invalid chars space", "aB3x 9z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := valueobject.ValidateHash(tt.hash); got != tt.want {
				t.Errorf("ValidateHash() = %v, want %v", got, tt.want)
			}
		})
	}
}
