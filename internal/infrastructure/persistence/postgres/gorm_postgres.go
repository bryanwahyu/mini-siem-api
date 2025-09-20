package postgres

import (
    "context"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"

    "server-analyst/internal/domain/entities"
    "server-analyst/internal/domain/repositories"
)

type DB struct{ *gorm.DB }

func Open(dsn string) (*DB, error) {
    gdb, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil { return nil, err }
    if err := gdb.AutoMigrate(&entities.Event{}, &entities.Detection{}, &entities.Decision{}, &entities.Rule{}, &entities.SpoolItem{}); err != nil { return nil, err }
    return &DB{gdb}, nil
}

type EventRepo struct{ db *gorm.DB }
type DecisionRepo struct{ db *gorm.DB }
type DetectionSaver struct{ db *gorm.DB }
type DetectionRepo struct{ db *gorm.DB }
type SpoolRepo struct{ db *gorm.DB }

func NewEventRepo(db *gorm.DB) *EventRepo { return &EventRepo{db: db} }
func NewDecisionRepo(db *gorm.DB) *DecisionRepo { return &DecisionRepo{db: db} }
func NewDetectionSaver(db *gorm.DB) *DetectionSaver { return &DetectionSaver{db: db} }
func NewDetectionRepo(db *gorm.DB) *DetectionRepo { return &DetectionRepo{db: db} }
func NewSpoolRepo(db *gorm.DB) *SpoolRepo { return &SpoolRepo{db: db} }

var _ repositories.EventRepository = (*EventRepo)(nil)
var _ repositories.DecisionRepository = (*DecisionRepo)(nil)
var _ repositories.DetectionRepository = (*DetectionRepo)(nil)
var _ repositories.SpoolRepository = (*SpoolRepo)(nil)

func (r *EventRepo) Save(ctx context.Context, ev *entities.Event) error {
    return r.db.WithContext(ctx).Create(ev).Error
}
func (r *EventRepo) List(ctx context.Context, limit, offset int) ([]entities.Event, error) {
    var out []entities.Event
    if err := r.db.WithContext(ctx).Order("id desc").Limit(limit).Offset(offset).Find(&out).Error; err != nil { return nil, err }
    return out, nil
}
func (r *EventRepo) Count(ctx context.Context) (int64, error) {
    var n int64
    if err := r.db.WithContext(ctx).Model(&entities.Event{}).Count(&n).Error; err != nil { return 0, err }
    return n, nil
}

func (r *DecisionRepo) Save(ctx context.Context, d *entities.Decision) error {
    return r.db.WithContext(ctx).Create(d).Error
}
func (r *DecisionRepo) List(ctx context.Context, limit, offset int) ([]entities.Decision, error) {
    var out []entities.Decision
    if err := r.db.WithContext(ctx).Order("id desc").Limit(limit).Offset(offset).Find(&out).Error; err != nil { return nil, err }
    return out, nil
}
func (r *DecisionRepo) Count(ctx context.Context) (int64, error) {
    var n int64
    if err := r.db.WithContext(ctx).Model(&entities.Decision{}).Count(&n).Error; err != nil { return 0, err }
    return n, nil
}

func (s *DetectionSaver) Save(ctx context.Context, d *entities.Detection) error {
    return s.db.WithContext(ctx).Create(d).Error
}

func (r *DetectionRepo) Save(ctx context.Context, d *entities.Detection) error {
    return r.db.WithContext(ctx).Create(d).Error
}
func (r *DetectionRepo) List(ctx context.Context, limit, offset int) ([]entities.Detection, error) {
    var out []entities.Detection
    if err := r.db.WithContext(ctx).Order("id desc").Limit(limit).Offset(offset).Find(&out).Error; err != nil { return nil, err }
    return out, nil
}
func (r *DetectionRepo) ListFiltered(ctx context.Context, f repositories.DetectionFilter, limit, offset int) ([]entities.Detection, error) {
    var out []entities.Detection
    db := r.db.WithContext(ctx).Table("detections").Select("detections.*").Joins("JOIN events ON events.id = detections.event_id")
    if f.IP != "" { db = db.Where("events.ip = ?", f.IP) }
    if f.Host != "" { db = db.Where("events.host = ?", f.Host) }
    if f.Source != "" { db = db.Where("events.source = ?", f.Source) }
    if f.Category != "" { db = db.Where("detections.category = ?", f.Category) }
    if f.Rule != "" { db = db.Where("detections.rule = ?", f.Rule) }
    if f.From != nil { db = db.Where("detections.created_at >= ?", *f.From) }
    if f.To != nil { db = db.Where("detections.created_at <= ?", *f.To) }
    if err := db.Order("detections.id desc").Limit(limit).Offset(offset).Scan(&out).Error; err != nil { return nil, err }
    return out, nil
}
func (r *DetectionRepo) Count(ctx context.Context) (int64, error) {
    var n int64
    if err := r.db.WithContext(ctx).Model(&entities.Detection{}).Count(&n).Error; err != nil { return 0, err }
    return n, nil
}

// SpoolRepo impl
func (r *SpoolRepo) Save(ctx context.Context, it *entities.SpoolItem) error {
    return r.db.WithContext(ctx).Create(it).Error
}
func (r *SpoolRepo) DeleteByID(ctx context.Context, id string) error {
    return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entities.SpoolItem{}).Error
}
