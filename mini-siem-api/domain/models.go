package domain

import "time"

// SeverityLevel represents the risk level associated with an event or detection.
type SeverityLevel string

const (
	SeverityLow      SeverityLevel = "low"
	SeverityMedium   SeverityLevel = "medium"
	SeverityHigh     SeverityLevel = "high"
	SeverityCritical SeverityLevel = "critical"
)

// Event represents an ingested security event/log entry.
type Event struct {
	ID         uint64
	ExternalID string
	Source     string
	IP         string
	Message    string
	Severity   SeverityLevel
	Timestamp  time.Time
	Metadata   map[string]any
	CreatedAt  time.Time
}

// Rule defines a detection rule evaluated against incoming events.
type Rule struct {
	ID          uint64
	Name        string
	Pattern     string
	Description string
	Severity    SeverityLevel
	Tags        []string
	Active      bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// Detection captures a rule match for a given event.
type Detection struct {
	ID        uint64
	EventID   uint64
	RuleID    uint64
	Summary   string
	Severity  SeverityLevel
	MatchedAt time.Time
	Metadata  map[string]any
}

// DecisionAction enumerates supported analyst actions.
type DecisionAction string

const (
	DecisionBlock   DecisionAction = "block"
	DecisionIgnore  DecisionAction = "ignore"
	DecisionMonitor DecisionAction = "monitor"
)

// Decision records a follow-up action taken for a detection or entity.
type Decision struct {
	ID          uint64
	DetectionID uint64
	Action      DecisionAction
	Target      string
	Reason      string
	Notes       string
	CreatedAt   time.Time
}
