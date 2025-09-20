package handlers

import (
	"context"
	"net/http"
	"time"

	"gorm.io/gorm"

	"server-analyst/mini-siem-api/infra/storage"
)

// Health godoc
// @Summary Service health check
// @Tags health
// @Produce json
// @Success 200 {object} map[string]any
// @Failure 503 {object} map[string]any
// @Router /health [get]
func Health(db *gorm.DB, archiver storage.Archiver) http.HandlerFunc {
	type healthArchiver interface {
		HealthCheck(ctx context.Context) error
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		result := map[string]any{"status": "ok"}
		statusCode := http.StatusOK

		if db != nil {
			sqlDB, err := db.DB()
			if err != nil || sqlDB.PingContext(ctx) != nil {
				result["database"] = "error"
				statusCode = http.StatusServiceUnavailable
			} else {
				result["database"] = "ok"
			}
		}

		if archiver != nil {
			if checker, ok := archiver.(healthArchiver); ok {
				if err := checker.HealthCheck(ctx); err != nil {
					result["archiver"] = "error"
					statusCode = http.StatusServiceUnavailable
				} else {
					result["archiver"] = "ok"
				}
			}
		}

		writeJSON(w, statusCode, result)
	}
}
