package entity_test

import (
	"testing"
	"time"

	"github.com/leonardo-gorska/nexuslink/internal/domain/entity"
)

func TestLink_IsExpired(t *testing.T) {
	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	tests := []struct {
		name      string
		expiresAt *time.Time
		want      bool
	}{
		{
			name:      "should not be expired if ExpiresAt is nil",
			expiresAt: nil,
			want:      false,
		},
		{
			name:      "should not be expired if ExpiresAt is in the future",
			expiresAt: &future,
			want:      false,
		},
		{
			name:      "should be expired if ExpiresAt is in the past",
			expiresAt: &past,
			want:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := &entity.Link{
				ExpiresAt: tt.expiresAt,
			}
			if got := l.IsExpired(); got != tt.want {
				t.Errorf("Link.IsExpired() = %v, want %v", got, tt.want)
			}
		})
	}
}
