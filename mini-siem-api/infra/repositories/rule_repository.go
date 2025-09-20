package repositories

import (
	"context"

	"gorm.io/gorm"

	"server-analyst/mini-siem-api/domain"
)

// RuleRepository persists detection rules.
type RuleRepository struct {
	db *gorm.DB
}

// NewRuleRepository constructs a rule repository.
func NewRuleRepository(db *gorm.DB) *RuleRepository {
	return &RuleRepository{db: db}
}

// Create inserts a new rule into the database.
func (r *RuleRepository) Create(ctx context.Context, rule *domain.Rule) error {
	model := toRuleModel(rule)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	rule.ID = model.ID
	rule.CreatedAt = model.CreatedAt
	rule.UpdatedAt = model.UpdatedAt
	return nil
}

// Update modifies a rule definition.
func (r *RuleRepository) Update(ctx context.Context, rule *domain.Rule) error {
	updates := map[string]any{
		"name":        rule.Name,
		"pattern":     rule.Pattern,
		"description": rule.Description,
		"severity":    string(rule.Severity),
		"tags":        rule.Tags,
		"active":      rule.Active,
	}
	return r.db.WithContext(ctx).
		Model(&RuleModel{}).
		Where("id = ?", rule.ID).
		Updates(updates).Error
}

// UpdateActivation toggles the active state of a rule.
func (r *RuleRepository) UpdateActivation(ctx context.Context, id uint64, active bool) error {
	return r.db.WithContext(ctx).Model(&RuleModel{}).Where("id = ?", id).Update("active", active).Error
}

// Get returns a single rule by ID.
func (r *RuleRepository) Get(ctx context.Context, id uint64) (*domain.Rule, error) {
	var model RuleModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return toDomainRule(&model), nil
}

// List returns all rules.
func (r *RuleRepository) List(ctx context.Context) ([]*domain.Rule, error) {
	var models []RuleModel
	if err := r.db.WithContext(ctx).Order("id ASC").Find(&models).Error; err != nil {
		return nil, err
	}
	rules := make([]*domain.Rule, 0, len(models))
	for i := range models {
		rules = append(rules, toDomainRule(&models[i]))
	}
	return rules, nil
}
