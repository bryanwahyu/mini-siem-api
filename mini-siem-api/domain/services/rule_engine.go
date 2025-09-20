package services

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"server-analyst/mini-siem-api/domain"
)

// RuleEngine performs regex-based rule evaluation against events.
type RuleEngine struct {
	mu    sync.RWMutex
	rules map[uint64]*compiledRule
}

type compiledRule struct {
	rule  *domain.Rule
	regex *regexp.Regexp
}

// NewRuleEngine initialises an empty rule engine.
func NewRuleEngine() *RuleEngine {
	return &RuleEngine{rules: make(map[uint64]*compiledRule)}
}

// Load replaces the rule set, compiling regex patterns in advance.
func (e *RuleEngine) Load(rules []*domain.Rule) error {
	compiled := make(map[uint64]*compiledRule, len(rules))
	for _, r := range rules {
		if !r.Active {
			continue
		}
		rx, err := regexp.Compile(r.Pattern)
		if err != nil {
			return fmt.Errorf("compile rule %d: %w", r.ID, err)
		}
		compiled[r.ID] = &compiledRule{rule: r, regex: rx}
	}

	e.mu.Lock()
	e.rules = compiled
	e.mu.Unlock()
	return nil
}

// Upsert updates or inserts a single rule in the engine.
func (e *RuleEngine) Upsert(rule *domain.Rule) error {
	if !rule.Active {
		e.Delete(rule.ID)
		return nil
	}
	rx, err := regexp.Compile(rule.Pattern)
	if err != nil {
		return err
	}
	e.mu.Lock()
	e.rules[rule.ID] = &compiledRule{rule: rule, regex: rx}
	e.mu.Unlock()
	return nil
}

// Delete removes a rule from evaluation.
func (e *RuleEngine) Delete(id uint64) {
	e.mu.Lock()
	delete(e.rules, id)
	e.mu.Unlock()
}

// Evaluate matches the incoming event against active rules.
func (e *RuleEngine) Evaluate(event *domain.Event) []*domain.Detection {
	now := time.Now().UTC()

	e.mu.RLock()
	defer e.mu.RUnlock()

	var detections []*domain.Detection
	for _, cr := range e.rules {
		if cr.regex.MatchString(event.Message) {
			detections = append(detections, &domain.Detection{
				RuleID:    cr.rule.ID,
				EventID:   event.ID,
				Severity:  cr.rule.Severity,
				Summary:   fmt.Sprintf("rule '%s' matched event", cr.rule.Name),
				MatchedAt: now,
				Metadata: map[string]any{
					"pattern":  cr.rule.Pattern,
					"ruleTags": cr.rule.Tags,
				},
			})
		}
	}
	return detections
}
