package repositories

import (
	"context"

	"gorm.io/gorm"

	"server-analyst/mini-siem-api/domain"
)

// DecisionRepository persists analyst decisions.
type DecisionRepository struct {
	db *gorm.DB
}

// NewDecisionRepository constructs the repository.
func NewDecisionRepository(db *gorm.DB) *DecisionRepository {
	return &DecisionRepository{db: db}
}

// Create stores a new decision record.
func (r *DecisionRepository) Create(ctx context.Context, decision *domain.Decision) error {
	model := toDecisionModel(decision)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	decision.ID = model.ID
	decision.CreatedAt = model.CreatedAt
	return nil
}

// List returns decisions with pagination.
func (r *DecisionRepository) List(ctx context.Context, filter domain.DecisionFilter) ([]*domain.Decision, error) {
	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query := r.db.WithContext(ctx).
		Model(&DecisionModel{}).
		Order("created_at DESC").
		Limit(limit).
		Offset(offset)

	var models []DecisionModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	decisions := make([]*domain.Decision, 0, len(models))
	for i := range models {
		decisions = append(decisions, toDomainDecision(&models[i]))
	}
	return decisions, nil
}
