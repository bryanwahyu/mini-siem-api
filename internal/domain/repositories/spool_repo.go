package repositories

import (
    "context"
    "server-analyst/internal/domain/entities"
)

type SpoolRepository interface {
    Save(ctx context.Context, it *entities.SpoolItem) error
    DeleteByID(ctx context.Context, id string) error
}

