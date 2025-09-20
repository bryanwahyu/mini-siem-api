package middleware

import (
	"net"
	"net/http"
	"strings"

	"github.com/rs/zerolog"

	"server-analyst/mini-siem-api/infra/ratelimit"
)

// RateLimiter enforces rate limits based on client IP.
func RateLimiter(limiter *ratelimit.Limiter, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if limiter == nil {
				next.ServeHTTP(w, r)
				return
			}
			key := clientIP(r)
			if !limiter.Allow(key) {
				logger.Warn().Str("ip", key).Msg("rate limit exceeded")
				http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func clientIP(r *http.Request) string {
	if header := r.Header.Get("X-Forwarded-For"); header != "" {
		parts := strings.Split(header, ",")
		ip := strings.TrimSpace(parts[0])
		if ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
