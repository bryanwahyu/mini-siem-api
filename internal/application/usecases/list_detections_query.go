package usecases

import (
    "context"
    "server-analyst/internal/domain/entities"
    "server-analyst/internal/domain/repositories"
)

type ListDetectionsQuery struct { Repo repositories.DetectionRepository }

func (q *ListDetectionsQuery) Run(ctx context.Context, filter *repositories.DetectionFilter, limit, offset int) ([]entities.Detection, error) {
    if filter != nil {
        return q.Repo.ListFiltered(ctx, *filter, limit, offset)
    }
    return q.Repo.List(ctx, limit, offset)
}
