package middleware

import (
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

// RequestLogger logs basic request metadata in structured form and injects logger into the context.
func RequestLogger(logger zerolog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqLogger := logger.With().Str("method", r.Method).Str("path", r.URL.Path).Logger()
			ctx := reqLogger.WithContext(r.Context())

			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
			next.ServeHTTP(rw, r.WithContext(ctx))

			duration := time.Since(start)
			reqLogger.Info().Int("status", rw.status).Dur("duration", duration).Msg("http request")
		})
	}
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.status = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
