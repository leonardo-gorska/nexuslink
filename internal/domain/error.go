package domain

import "errors"

var (
	ErrLinkNotFound  = errors.New("link not found or inactive")
	ErrInvalidURL    = errors.New("invalid URL format")
	ErrHashCollision = errors.New("hash collision occurred after max retries")
	ErrRateLimited   = errors.New("rate limit exceeded")
	ErrLinkExpired   = errors.New("link has expired")
	ErrCacheMiss     = errors.New("cache miss")
)
