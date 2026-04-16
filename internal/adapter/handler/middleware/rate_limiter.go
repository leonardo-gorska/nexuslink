package middleware

import (
	"context"
	"net"
	"net/http"
	"time"

	"github.com/leonardo-gorska/nexuslink/pkg/httputil"
)

type RateLimiterStore interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

func RateLimiter(store RateLimiterStore, limit int, window time.Duration) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.Header.Get("X-Real-IP")
			if ip == "" {
				ip = r.Header.Get("X-Forwarded-For")
			}
			if ip == "" {
				var err error
				ip, _, err = net.SplitHostPort(r.RemoteAddr)
				if err != nil {
					ip = r.RemoteAddr
				}
			}

			key := "rl:" + ip

			allowed, err := store.Allow(r.Context(), key, limit, window)
			if err != nil {
				// Degrade gracefully if redis is down
				next.ServeHTTP(w, r)
				return
			}

			if !allowed {
				reqID := GetReqID(r.Context())
				w.Header().Set("Retry-After", "60")
				httputil.WriteProblem(w, httputil.ProblemDetail{
					Type:      "https://nexuslink.dev/errors/rate-limited",
					Title:     "Rate Limit Exceeded",
					Status:    http.StatusTooManyRequests,
					Detail:    "You have exceeded your rate limit.",
					Instance:  r.URL.Path,
					RequestID: reqID,
				})
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
