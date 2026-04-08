package rate_limiter

import (
	"log/slog"
	"net/http"

	"github.com/avito-internships/test-backend-1-mrpixik/internal/adapters/logger"
)

type rateLimiter interface {
	Allow() bool
}

func WithRateLimiter(limiter rateLimiter, log logger.LoggerAdapter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		log = log.With(
			slog.String("component", "middleware/rate_limiter"),
		)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				log.Info("rate limit exceeded",
					slog.String("method", r.Method),
					slog.String("url", r.URL.String()),
				)

				w.Header().Set("X-RateLimit-Limit", "10")
				w.Header().Set("X-RateLimit-Remaining", "0")
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte("rate limit exceeded"))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
