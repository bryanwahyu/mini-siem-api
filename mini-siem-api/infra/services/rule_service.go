package services

import (
	"context"

	"github.com/rs/zerolog"

	"server-analyst/mini-siem-api/domain"
)

// RuleService manages rule lifecycle and keeps the engine in sync.
type RuleService struct {
	repo         domain.RuleRepository
	eventService *EventService
	logger       zerolog.Logger
}

// NewRuleService constructs the rule management service.
func NewRuleService(repo domain.RuleRepository, eventService *EventService, logger zerolog.Logger) *RuleService {
	return &RuleService{repo: repo, eventService: eventService, logger: logger}
}

// List returns all rules.
func (s *RuleService) List(ctx context.Context) ([]*domain.Rule, error) {
	return s.repo.List(ctx)
}

// Create persists a new rule and reloads the engine.
func (s *RuleService) Create(ctx context.Context, rule *domain.Rule) error {
	if err := s.repo.Create(ctx, rule); err != nil {
		return err
	}
	s.logger.Info().Uint64("rule_id", rule.ID).Msg("rule created")
	return s.eventService.ReloadRules(ctx)
}

// SetActivation toggles the rule active flag.
func (s *RuleService) SetActivation(ctx context.Context, id uint64, active bool) error {
	if err := s.repo.UpdateActivation(ctx, id, active); err != nil {
		return err
	}
	s.logger.Info().Uint64("rule_id", id).Bool("active", active).Msg("rule activation updated")
	return s.eventService.ReloadRules(ctx)
}
