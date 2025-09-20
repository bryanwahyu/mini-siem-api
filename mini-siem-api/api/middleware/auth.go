package middleware

import (
	"net/http"
	"strings"

	"github.com/rs/zerolog"
)

// APIKeyAuth enforces the presence of a static API key in admin endpoints.
func APIKeyAuth(expectedKey string, logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if expectedKey == "" {
				logger.Warn().Msg("admin endpoint accessed without API key configured")
				next.ServeHTTP(w, r)
				return
			}
			key := r.Header.Get("X-API-Key")
			if key == "" {
				key = r.URL.Query().Get("api_key")
			}
			if strings.TrimSpace(key) != expectedKey {
				http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
