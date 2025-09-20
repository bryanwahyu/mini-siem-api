package repositories

import (
    "context"
    "server-analyst/internal/domain/entities"
)

type DecisionRepository interface {
    Save(ctx context.Context, d *entities.Decision) error
    List(ctx context.Context, limit, offset int) ([]entities.Decision, error)
    Count(ctx context.Context) (int64, error)
}
