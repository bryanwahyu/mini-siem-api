package repositories

import (
    "context"
    "server-analyst/internal/domain/entities"
)

type RuleRepository interface {
    Save(ctx context.Context, r *entities.Rule) error
    List(ctx context.Context, onlyEnabled bool) ([]entities.Rule, error)
}
