package repositories

import (
	"context"
	"time"

	"gorm.io/gorm"

	"server-analyst/mini-siem-api/domain"
)

// EventRepository implements domain.EventRepository using GORM.
type EventRepository struct {
	db *gorm.DB
}

// NewEventRepository creates a new repository instance.
func NewEventRepository(db *gorm.DB) *EventRepository {
	return &EventRepository{db: db}
}

// Save persists a single event.
func (r *EventRepository) Save(ctx context.Context, event *domain.Event) error {
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now().UTC()
	}
	model := toEventModel(event)
	if err := r.db.WithContext(ctx).Create(model).Error; err != nil {
		return err
	}
	event.ID = model.ID
	event.CreatedAt = model.CreatedAt
	return nil
}

// SaveBatch persists multiple events in a transaction.
func (r *EventRepository) SaveBatch(ctx context.Context, events []*domain.Event) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for _, evt := range events {
			if evt.Timestamp.IsZero() {
				evt.Timestamp = time.Now().UTC()
			}
			model := toEventModel(evt)
			if err := tx.Create(model).Error; err != nil {
				return err
			}
			evt.ID = model.ID
			evt.CreatedAt = model.CreatedAt
		}
		return nil
	})
}

// List returns events applying optional filters.
func (r *EventRepository) List(ctx context.Context, filter domain.EventFilter) ([]*domain.Event, error) {
	query := r.db.WithContext(ctx).Model(&EventModel{}).Order("created_at DESC")

	if filter.Source != "" {
		query = query.Where("source = ?", filter.Source)
	}
	if filter.IP != "" {
		query = query.Where("ip = ?", filter.IP)
	}
	if filter.From != nil {
		query = query.Where("timestamp >= ?", filter.From)
	}
	if filter.To != nil {
		query = query.Where("timestamp <= ?", filter.To)
	}

	limit := filter.Limit
	if limit <= 0 || limit > 500 {
		limit = 50
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	query = query.Limit(limit).Offset(offset)

	var models []EventModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	events := make([]*domain.Event, 0, len(models))
	for i := range models {
		events = append(events, toDomainEvent(&models[i]))
	}
	return events, nil
}

// FindByID returns a single event by its identifier.
func (r *EventRepository) FindByID(ctx context.Context, id uint64) (*domain.Event, error) {
	var model EventModel
	if err := r.db.WithContext(ctx).First(&model, id).Error; err != nil {
		return nil, err
	}
	return toDomainEvent(&model), nil
}
