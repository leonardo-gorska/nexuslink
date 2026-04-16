package entity

import "time"

// ClickEvent represents a raw click event on a short link.
type ClickEvent struct {
	ID         int64
	LinkHash   string
	IP         string
	UserAgent  string
	Referer    string
	Country    string
	Device     string
	Browser    string
	OS         string
	ClickedAt  time.Time
}
