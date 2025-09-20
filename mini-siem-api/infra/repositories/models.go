package repositories

import (
	"time"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// EventModel maps the domain event into persistent storage.
type EventModel struct {
	ID         uint64 `gorm:"primaryKey"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
	ExternalID string            `gorm:"size:128;index"`
	Source     string            `gorm:"size:64;index"`
	IP         string            `gorm:"size:64;index"`
	Message    string            `gorm:"type:text"`
	Severity   string            `gorm:"size:16;index"`
	Timestamp  time.Time         `gorm:"index"`
	Metadata   datatypes.JSONMap `gorm:"serializer:json"`
}

func (EventModel) TableName() string { return "events" }

// RuleModel holds detection rule definitions.
type RuleModel struct {
	ID          uint64 `gorm:"primaryKey"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string                      `gorm:"size:128;uniqueIndex"`
	Pattern     string                      `gorm:"type:text"`
	Description string                      `gorm:"type:text"`
	Severity    string                      `gorm:"size:16"`
	Tags        datatypes.JSONSlice[string] `gorm:"serializer:json"`
	Active      bool                        `gorm:"index"`
}

func (RuleModel) TableName() string { return "rules" }

// DetectionModel stores rule matches for events.
type DetectionModel struct {
	ID        uint64 `gorm:"primaryKey"`
	CreatedAt time.Time
	EventID   uint64            `gorm:"index"`
	RuleID    uint64            `gorm:"index"`
	Summary   string            `gorm:"type:text"`
	Severity  string            `gorm:"size:16;index"`
	MatchedAt time.Time         `gorm:"index"`
	Metadata  datatypes.JSONMap `gorm:"serializer:json"`
}

func (DetectionModel) TableName() string { return "detections" }

// DecisionModel keeps analyst actions.
type DecisionModel struct {
	ID          uint64 `gorm:"primaryKey"`
	CreatedAt   time.Time
	DetectionID uint64 `gorm:"index"`
	Action      string `gorm:"size:32;index"`
	Target      string `gorm:"size:128;index"`
	Reason      string `gorm:"type:text"`
	Notes       string `gorm:"type:text"`
}

func (DecisionModel) TableName() string { return "decisions" }

// AutoMigrate runs the schema migrations for the mini SIEM entities.
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(&EventModel{}, &RuleModel{}, &DetectionModel{}, &DecisionModel{})
}
