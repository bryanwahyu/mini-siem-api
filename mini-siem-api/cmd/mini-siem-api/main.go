package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"

	"server-analyst/mini-siem-api/api"
	docs "server-analyst/mini-siem-api/api/docs"
	"server-analyst/mini-siem-api/config"
	"server-analyst/mini-siem-api/infra/database"
	"server-analyst/mini-siem-api/infra/metrics"
	"server-analyst/mini-siem-api/infra/repositories"
	"server-analyst/mini-siem-api/infra/services"
)

func main() {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	cfg, err := config.Load()
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to load config")
	}

	docs.SwaggerInfo.Title = "Mini SIEM API"
	docs.SwaggerInfo.Description = "Rule-based anomaly detection API with event ingestion, detections, and analyst decisions"
	docs.SwaggerInfo.Version = "1.0"
	docs.SwaggerInfo.BasePath = "/"

	db, err := database.Open(cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect database")
	}
	defer func() {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}()

	if err := repositories.AutoMigrate(db); err != nil {
		logger.Fatal().Err(err).Msg("failed to migrate database")
	}

	promRegistry := prometheus.DefaultRegisterer
	metricCollector := metrics.New(promRegistry)
	archiver := cfg.MinIOArchiver()

	eventRepo := repositories.NewEventRepository(db)
	detectionRepo := repositories.NewDetectionRepository(db)
	ruleRepo := repositories.NewRuleRepository(db)
	decisionRepo := repositories.NewDecisionRepository(db)

	ctx := context.Background()
	eventService, err := services.NewEventService(ctx, eventRepo, detectionRepo, ruleRepo, archiver, metricCollector, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to initialise event service")
	}

	ruleService := services.NewRuleService(ruleRepo, eventService, logger)
	decisionService := services.NewDecisionService(decisionRepo, detectionRepo, logger)

	deps := api.Dependencies{
		EventService:    eventService,
		RuleService:     ruleService,
		DecisionService: decisionService,
		EventRepo:       eventRepo,
		DetectionRepo:   detectionRepo,
		Logger:          logger,
		RateLimiter:     cfg.BuildLimiter(),
		AdminAPIKey:     cfg.Server.AdminAPIKey,
		Metrics:         metricCollector,
		DB:              db,
		Archiver:        archiver,
	}

	router := api.NewRouter(deps)

	srv := &http.Server{
		Addr:         cfg.Server.Address,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logger.Info().Str("addr", cfg.Server.Address).Msg("mini SIEM API started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("http server crashed")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	logger.Info().Msg("shutting down mini SIEM API")
	if err := srv.Shutdown(ctxShutdown); err != nil {
		logger.Error().Err(err).Msg("graceful shutdown failed")
	}
}
