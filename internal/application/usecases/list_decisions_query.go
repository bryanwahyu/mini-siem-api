package usecases

import (
    "context"
    "server-analyst/internal/domain/entities"
    "server-analyst/internal/domain/repositories"
)

type ListDecisionsQuery struct { Repo repositories.DecisionRepository }

func (q *ListDecisionsQuery) Run(ctx context.Context, limit, offset int) ([]entities.Decision, error) {
    return q.Repo.List(ctx, limit, offset)
}
