package services

import (
	"context"

	"github.com/rs/zerolog"

	"server-analyst/mini-siem-api/domain"
)

// DecisionService coordinates analyst decisions persistence.
type DecisionService struct {
	repo       domain.DecisionRepository
	detections domain.DetectionRepository
	logger     zerolog.Logger
}

// NewDecisionService constructs a decision service.
func NewDecisionService(repo domain.DecisionRepository, detections domain.DetectionRepository, logger zerolog.Logger) *DecisionService {
	return &DecisionService{repo: repo, detections: detections, logger: logger}
}

// Create stores a decision after validating the detection exists.
func (s *DecisionService) Create(ctx context.Context, decision *domain.Decision) error {
	if decision.DetectionID != 0 {
		if _, err := s.detections.Get(ctx, decision.DetectionID); err != nil {
			return err
		}
	}
	if err := s.repo.Create(ctx, decision); err != nil {
		return err
	}
	s.logger.Info().Uint64("decision_id", decision.ID).Str("action", string(decision.Action)).Msg("decision recorded")
	return nil
}

// List returns stored decisions.
func (s *DecisionService) List(ctx context.Context, filter domain.DecisionFilter) ([]*domain.Decision, error) {
	return s.repo.List(ctx, filter)
}
