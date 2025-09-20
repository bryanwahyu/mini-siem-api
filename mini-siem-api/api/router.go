package api

import (
	"github.com/go-chi/chi/v5"
	stdmiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/zerolog"
	httpSwagger "github.com/swaggo/http-swagger"
	"gorm.io/gorm"

	httpi "server-analyst/mini-siem-api/api/handlers"
	mid "server-analyst/mini-siem-api/api/middleware"
	"server-analyst/mini-siem-api/domain"
	"server-analyst/mini-siem-api/infra/metrics"
	"server-analyst/mini-siem-api/infra/ratelimit"
	"server-analyst/mini-siem-api/infra/services"
	"server-analyst/mini-siem-api/infra/storage"
)

// Dependencies contains the runtime wiring for the HTTP API.
type Dependencies struct {
	EventService    *services.EventService
	RuleService     *services.RuleService
	DecisionService *services.DecisionService
	EventRepo       domain.EventRepository
	DetectionRepo   domain.DetectionRepository
	Logger          zerolog.Logger
	RateLimiter     *ratelimit.Limiter
	AdminAPIKey     string
	Metrics         *metrics.Metrics
	DB              *gorm.DB
	Archiver        storage.Archiver
}

// NewRouter builds the HTTP router with all middleware and routes.
func NewRouter(deps Dependencies) chi.Router {
	app := chi.NewRouter()

	app.Use(stdmiddleware.RequestID)
	app.Use(stdmiddleware.RealIP)
	app.Use(stdmiddleware.Recoverer)
	app.Use(mid.RequestLogger(deps.Logger))

	// Observability & docs
	app.Get("/health", httpi.Health(deps.DB, deps.Archiver))
	app.Get("/metrics", promhttp.Handler().ServeHTTP)
	app.Get("/swagger/*", httpSwagger.WrapHandler)

	// Event ingestion with rate limiting
	app.With(mid.RateLimiter(deps.RateLimiter, deps.Logger)).Post("/events", httpi.PostEvents(deps.EventService))

	// Event queries
	app.Get("/events", httpi.GetEvents(deps.EventRepo))
	app.Get("/events/{id}", httpi.GetEventByID(deps.EventRepo))

	// Detections
	app.Get("/detections", httpi.GetDetections(deps.DetectionRepo))
	app.Get("/detections/{id}", httpi.GetDetectionByID(deps.DetectionRepo))

	// Admin endpoints secured via API key
	admin := app.With(mid.APIKeyAuth(deps.AdminAPIKey, deps.Logger))
	admin.Get("/rules", httpi.GetRules(deps.RuleService))
	admin.Post("/rules", httpi.PostRules(deps.RuleService))
	admin.Patch("/rules/{id}", httpi.PatchRuleActivation(deps.RuleService))
	admin.Get("/decisions", httpi.GetDecisions(deps.DecisionService))
	admin.Post("/decisions", httpi.PostDecisions(deps.DecisionService))

	return app
}
