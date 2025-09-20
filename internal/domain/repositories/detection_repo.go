package repositories

import (
    "context"
    "time"
    "server-analyst/internal/domain/entities"
)

type DetectionFilter struct {
    IP       string
    Host     string
    Source   string
    Category string
    Rule     string
    From     *time.Time
    To       *time.Time
}

type DetectionRepository interface {
    Save(ctx context.Context, d *entities.Detection) error
    List(ctx context.Context, limit, offset int) ([]entities.Detection, error)
    ListFiltered(ctx context.Context, f DetectionFilter, limit, offset int) ([]entities.Detection, error)
    Count(ctx context.Context) (int64, error)
}
