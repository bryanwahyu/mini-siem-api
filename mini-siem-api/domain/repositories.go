package domain

import (
	"context"
	"time"
)

// EventFilter holds query parameters for listing events.
type EventFilter struct {
	Source string
	IP     string
	From   *time.Time
	To     *time.Time
	Limit  int
	Offset int
}

// EventRepository persists security events.
type EventRepository interface {
	Save(ctx context.Context, event *Event) error
	SaveBatch(ctx context.Context, events []*Event) error
	List(ctx context.Context, filter EventFilter) ([]*Event, error)
	FindByID(ctx context.Context, id uint64) (*Event, error)
}

// RuleRepository manages detection rules.
type RuleRepository interface {
	Create(ctx context.Context, rule *Rule) error
	Update(ctx context.Context, rule *Rule) error
	UpdateActivation(ctx context.Context, id uint64, active bool) error
	Get(ctx context.Context, id uint64) (*Rule, error)
	List(ctx context.Context) ([]*Rule, error)
}

// DetectionFilter encapsulates filter params for detection listing.
type DetectionFilter struct {
	From     *time.Time
	To       *time.Time
	Severity []SeverityLevel
	RuleIDs  []uint64
	Limit    int
	Offset   int
}

// DetectionRepository persists rule matches.
type DetectionRepository interface {
	Create(ctx context.Context, detection *Detection) error
	BatchCreate(ctx context.Context, detections []*Detection) error
	Get(ctx context.Context, id uint64) (*Detection, error)
	List(ctx context.Context, filter DetectionFilter) ([]*Detection, error)
}

// DecisionFilter defines query options for decisions.
type DecisionFilter struct {
	Limit  int
	Offset int
}

// DecisionRepository persists analyst/operator decisions.
type DecisionRepository interface {
	Create(ctx context.Context, decision *Decision) error
	List(ctx context.Context, filter DecisionFilter) ([]*Decision, error)
}
