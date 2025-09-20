package services

import (
	"context"

	"github.com/rs/zerolog"

	"server-analyst/mini-siem-api/domain"
	domainservices "server-analyst/mini-siem-api/domain/services"
	"server-analyst/mini-siem-api/infra/metrics"
	"server-analyst/mini-siem-api/infra/storage"
)

// EventService orchestrates event ingestion, detection, and archiving.
type EventService struct {
	eventsRepo     domain.EventRepository
	detectionsRepo domain.DetectionRepository
	rulesRepo      domain.RuleRepository
	engine         *domainservices.RuleEngine
	archiver       storage.Archiver
	metrics        *metrics.Metrics
	logger         zerolog.Logger
}

// NewEventService constructs the service and warms the rule engine.
func NewEventService(
	ctx context.Context,
	events domain.EventRepository,
	detections domain.DetectionRepository,
	rules domain.RuleRepository,
	archiver storage.Archiver,
	metrics *metrics.Metrics,
	logger zerolog.Logger,
) (*EventService, error) {
	if archiver == nil {
		archiver = storage.NoopArchiver{}
	}
	engine := domainservices.NewRuleEngine()

	ruleList, err := rules.List(ctx)
	if err != nil {
		return nil, err
	}
	if err := engine.Load(ruleList); err != nil {
		return nil, err
	}

	return &EventService{
		eventsRepo:     events,
		detectionsRepo: detections,
		rulesRepo:      rules,
		engine:         engine,
		archiver:       archiver,
		metrics:        metrics,
		logger:         logger,
	}, nil
}

// ReloadRules refreshes the rule engine cache.
func (s *EventService) ReloadRules(ctx context.Context) error {
	rules, err := s.rulesRepo.List(ctx)
	if err != nil {
		return err
	}
	return s.engine.Load(rules)
}

// Ingest stores a single event and runs detection rules.
func (s *EventService) Ingest(ctx context.Context, event *domain.Event) ([]*domain.Detection, error) {
	if event == nil {
		return nil, nil
	}
	if err := s.eventsRepo.Save(ctx, event); err != nil {
		s.recordError()
		s.logger.Error().Err(err).Str("source", event.Source).Msg("failed to persist event")
		return nil, err
	}
	s.recordEvent()
	_ = s.archiver.ArchiveEvent(ctx, event) // best effort

	detections := s.engine.Evaluate(event)
	if err := s.persistDetections(ctx, detections); err != nil {
		s.recordError()
		s.logger.Error().Err(err).Uint64("event_id", event.ID).Msg("failed to persist detections")
		return nil, err
	}
	s.logger.Info().Uint64("event_id", event.ID).Int("detections", len(detections)).Msg("event ingested")
	return detections, nil
}

// IngestBatch stores multiple events and detects anomalies.
func (s *EventService) IngestBatch(ctx context.Context, events []*domain.Event) ([]*domain.Detection, error) {
	if len(events) == 0 {
		return nil, nil
	}
	if err := s.eventsRepo.SaveBatch(ctx, events); err != nil {
		s.recordError()
		s.logger.Error().Err(err).Int("events", len(events)).Msg("failed to persist batch events")
		return nil, err
	}

	var allDetections []*domain.Detection
	for _, evt := range events {
		s.recordEvent()
		_ = s.archiver.ArchiveEvent(ctx, evt)
		detections := s.engine.Evaluate(evt)
		allDetections = append(allDetections, detections...)
	}
	if err := s.persistDetections(ctx, allDetections); err != nil {
		s.recordError()
		s.logger.Error().Err(err).Int("events", len(events)).Msg("failed to persist batch detections")
		return nil, err
	}
	s.logger.Info().Int("events", len(events)).Int("detections", len(allDetections)).Msg("batch ingested")
	return allDetections, nil
}

func (s *EventService) persistDetections(ctx context.Context, detections []*domain.Detection) error {
	if len(detections) == 0 {
		return nil
	}
	if err := s.detectionsRepo.BatchCreate(ctx, detections); err != nil {
		return err
	}
	for _, det := range detections {
		s.recordDetection(string(det.Severity))
		_ = s.archiver.ArchiveDetection(ctx, det)
	}
	return nil
}

func (s *EventService) recordEvent() {
	if s.metrics != nil {
		s.metrics.RecordEvent()
	}
}

func (s *EventService) recordDetection(severity string) {
	if s.metrics != nil {
		s.metrics.RecordDetection(severity)
	}
}

func (s *EventService) recordError() {
	if s.metrics != nil {
		s.metrics.RecordError()
	}
}

// Rules returns all configured rules, reusing the repository.
func (s *EventService) Rules(ctx context.Context) ([]*domain.Rule, error) {
	return s.rulesRepo.List(ctx)
}

// RuleRepository exposes rule repository for consumers.
func (s *EventService) RuleRepository() domain.RuleRepository {
	return s.rulesRepo
}

// DetectionRepository exposes detection repo for consumers.
func (s *EventService) DetectionRepository() domain.DetectionRepository {
	return s.detectionsRepo
}

// EventRepository exposes event repo for consumers.
func (s *EventService) EventRepository() domain.EventRepository {
	return s.eventsRepo
}
