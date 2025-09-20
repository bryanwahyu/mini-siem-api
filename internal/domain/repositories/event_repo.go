package repositories

import (
    "context"
    "server-analyst/internal/domain/entities"
)

type EventRepository interface {
    Save(ctx context.Context, ev *entities.Event) error
    List(ctx context.Context, limit, offset int) ([]entities.Event, error)
    Count(ctx context.Context) (int64, error)
}
