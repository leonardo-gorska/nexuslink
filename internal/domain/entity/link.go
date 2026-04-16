package entity

import "time"

// Link represents a shortened URL.
type Link struct {
	ID          int64
	Hash        string
	OriginalURL string
	CreatedAt   time.Time
	ExpiresAt   *time.Time
	ClickCount  int64
	IsActive    bool
}

// IsExpired checks if the link has an expiration date and it has passed.
func (l *Link) IsExpired() bool {
	if l.ExpiresAt == nil {
		return false
	}
	return l.ExpiresAt.Before(time.Now())
}
