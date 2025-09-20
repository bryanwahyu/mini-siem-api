package handlers

import (
	"encoding/json"
	"errors"
	"time"

	"server-analyst/mini-siem-api/domain"
)

// EventRequest wraps incoming event payloads.
type EventRequest struct {
	ExternalID string         `json:"external_id"`
	Source     string         `json:"source"`
	IP         string         `json:"ip"`
	Message    string         `json:"message"`
	Severity   string         `json:"severity"`
	Timestamp  string         `json:"timestamp"`
	Metadata   map[string]any `json:"metadata"`
}

func (e EventRequest) ToDomain() (*domain.Event, error) {
	if e.Message == "" {
		return nil, errors.New("message is required")
	}
	severity, err := domain.ParseSeverity(e.Severity)
	if err != nil {
		return nil, err
	}
	ts := time.Now().UTC()
	if e.Timestamp != "" {
		parsed, err := time.Parse(time.RFC3339, e.Timestamp)
		if err != nil {
			return nil, err
		}
		ts = parsed
	}
	metadata := make(map[string]any)
	for k, v := range e.Metadata {
		metadata[k] = v
	}
	return &domain.Event{
		ExternalID: e.ExternalID,
		Source:     e.Source,
		IP:         e.IP,
		Message:    e.Message,
		Severity:   severity,
		Timestamp:  ts,
		Metadata:   metadata,
	}, nil
}

// EventBatchRequest handles either single or batch payloads.
type EventBatchRequest []EventRequest

func (e EventBatchRequest) ToDomain() ([]*domain.Event, error) {
	result := make([]*domain.Event, 0, len(e))
	for _, item := range e {
		evt, err := item.ToDomain()
		if err != nil {
			return nil, err
		}
		result = append(result, evt)
	}
	return result, nil
}

// RuleRequest describes a rule creation payload.
type RuleRequest struct {
	Name        string   `json:"name"`
	Pattern     string   `json:"pattern"`
	Description string   `json:"description"`
	Severity    string   `json:"severity"`
	Tags        []string `json:"tags"`
	Active      *bool    `json:"active"`
}

func (r RuleRequest) ToDomain() (*domain.Rule, error) {
	if r.Name == "" || r.Pattern == "" {
		return nil, errors.New("name and pattern are required")
	}
	severity, err := domain.ParseSeverity(r.Severity)
	if err != nil {
		return nil, err
	}
	active := true
	if r.Active != nil {
		active = *r.Active
	}
	tags := make([]string, len(r.Tags))
	copy(tags, r.Tags)
	return &domain.Rule{
		Name:        r.Name,
		Pattern:     r.Pattern,
		Description: r.Description,
		Severity:    severity,
		Tags:        tags,
		Active:      active,
	}, nil
}

// RuleActivationRequest toggles a rule.
type RuleActivationRequest struct {
	Active bool `json:"active"`
}

// DecisionRequest contains analyst decision payload.
type DecisionRequest struct {
	DetectionID uint64 `json:"detection_id"`
	Action      string `json:"action"`
	Target      string `json:"target"`
	Reason      string `json:"reason"`
	Notes       string `json:"notes"`
}

func (d DecisionRequest) ToDomain() (*domain.Decision, error) {
	action, err := domain.ParseDecisionAction(d.Action)
	if err != nil {
		return nil, err
	}
	return &domain.Decision{
		DetectionID: d.DetectionID,
		Action:      action,
		Target:      d.Target,
		Reason:      d.Reason,
		Notes:       d.Notes,
	}, nil
}

// DetectionResponse is returned to clients.
type DetectionResponse struct {
	ID        uint64         `json:"id"`
	EventID   uint64         `json:"event_id"`
	RuleID    uint64         `json:"rule_id"`
	Summary   string         `json:"summary"`
	Severity  string         `json:"severity"`
	MatchedAt time.Time      `json:"matched_at"`
	Metadata  map[string]any `json:"metadata"`
}

// RuleResponse exposes rule info.
type RuleResponse struct {
	ID          uint64    `json:"id"`
	Name        string    `json:"name"`
	Pattern     string    `json:"pattern"`
	Description string    `json:"description"`
	Severity    string    `json:"severity"`
	Tags        []string  `json:"tags"`
	Active      bool      `json:"active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DecisionResponse returns decision data.
type DecisionResponse struct {
	ID          uint64    `json:"id"`
	DetectionID uint64    `json:"detection_id"`
	Action      string    `json:"action"`
	Target      string    `json:"target"`
	Reason      string    `json:"reason"`
	Notes       string    `json:"notes"`
	CreatedAt   time.Time `json:"created_at"`
}

// EventResponse serialises stored events.
type EventResponse struct {
	ID        uint64         `json:"id"`
	Source    string         `json:"source"`
	IP        string         `json:"ip"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`
}

// ErrorResponse is used in error payloads.
type ErrorResponse struct {
	Error string `json:"error"`
}

// MarshalEvents attempts to decode payload either as single or array.
func MarshalEvents(raw json.RawMessage) (EventBatchRequest, error) {
	if len(raw) == 0 {
		return nil, errors.New("empty payload")
	}
	trimmed := bytesTrim(raw)
	if len(trimmed) == 0 {
		return nil, errors.New("empty payload")
	}
	if trimmed[0] == '{' {
		var single EventRequest
		if err := json.Unmarshal(raw, &single); err != nil {
			return nil, err
		}
		return EventBatchRequest{single}, nil
	}
	var batch EventBatchRequest
	if err := json.Unmarshal(raw, &batch); err != nil {
		return nil, err
	}
	return batch, nil
}

func bytesTrim(raw []byte) []byte {
	start := 0
	end := len(raw)
	for start < end && isWhitespace(raw[start]) {
		start++
	}
	for end > start && isWhitespace(raw[end-1]) {
		end--
	}
	return raw[start:end]
}

func isWhitespace(b byte) bool {
	switch b {
	case ' ', '\n', '\r', '\t':
		return true
	default:
		return false
	}
}
