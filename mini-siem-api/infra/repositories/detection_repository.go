package repositories

import (
	"context"

	"gorm.io/gorm"

	"server-analyst/mini-siem-api/domain"
)

// DetectionRepository persists detections using GORM.
type DetectionRepository struct {
	db *gorm.DB
}

// NewDetectionRepository constructs the repository.
func NewDetectionRepository(db *gorm.DB) *DetectionRepository {
	return &DetectionRepository{db: db}
}

// Create inserts a single detection.
func (r *DetectionRepository) Create(ctx context.Context, detection *domain.Detection) error {
	model := toDetectionModel(detection)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	detection.ID = model.ID
	detection.MatchedAt = model.MatchedAt
	return nil
}

// BatchCreate inserts multiple detections transactionally.
func (r *DetectionRepository) BatchCreate(ctx context.Context, detections []*domain.Detection) error {
	if len(detections) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, detection := range detections {
			model := toDetectionModel(detection)
			if err := tx.Create(model).Error; err != nil {
				return err
			}
			detection.ID = model.ID
			detection.MatchedAt = model.MatchedAt
		}
		return nil
	})
}

// Get fetches a detection by ID.
func (r *DetectionRepository) Get(ctx context.Context, id uint64) (*domain.Detection, error) {
	var model DetectionModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return toDomainDetection(&model), nil
}

// List returns detections for the supplied filter.
func (r *DetectionRepository) List(ctx context.Context, filter domain.DetectionFilter) ([]*domain.Detection, error) {
	query := r.db.WithContext(ctx).Model(&DetectionModel{}).Order("matched_at DESC")

	if filter.From != nil {
		query = query.Where("matched_at >= ?", filter.From)
	}
	if filter.To != nil {
		query = query.Where("matched_at <= ?", filter.To)
	}
	if len(filter.Severity) > 0 {
		severities := make([]string, len(filter.Severity))
		for i, s := range filter.Severity {
			severities[i] = string(s)
		}
		query = query.Where("severity IN ?", severities)
	}
	if len(filter.RuleIDs) > 0 {
		query = query.Where("rule_id IN ?", filter.RuleIDs)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 100
	}

	query = query.Limit(limit).Offset(filter.Offset)

	var models []DetectionModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	detections := make([]*domain.Detection, 0, len(models))
	for i := range models {
		detections = append(detections, toDomainDetection(&models[i]))
	}
	return detections, nil
}
