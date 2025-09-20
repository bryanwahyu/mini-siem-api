package usecases

import (
    "context"
    "server-analyst/internal/domain/entities"
    "server-analyst/internal/domain/repositories"
)

type ListEventsQuery struct { Repo repositories.EventRepository }

func (q *ListEventsQuery) Run(ctx context.Context, limit, offset int) ([]entities.Event, error) {
    return q.Repo.List(ctx, limit, offset)
}
