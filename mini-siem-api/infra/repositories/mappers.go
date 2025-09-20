package repositories

import (
	"time"

	"server-analyst/mini-siem-api/domain"
)

func toEventModel(event *domain.Event) *EventModel {
	metadata := make(map[string]any)
	if event.Metadata != nil {
		for k, v := range event.Metadata {
			metadata[k] = v
		}
	}
	return &EventModel{
		ID:         event.ID,
		ExternalID: event.ExternalID,
		Source:     event.Source,
		IP:         event.IP,
		Message:    event.Message,
		Severity:   string(event.Severity),
		Timestamp:  event.Timestamp,
		Metadata:   metadata,
	}
}

func toDomainEvent(model *EventModel) *domain.Event {
	if model == nil {
		return nil
	}
	md := make(map[string]any)
	for k, v := range model.Metadata {
		md[k] = v
	}
	return &domain.Event{
		ID:         model.ID,
		ExternalID: model.ExternalID,
		Source:     model.Source,
		IP:         model.IP,
		Message:    model.Message,
		Severity:   domain.SeverityLevel(model.Severity),
		Timestamp:  model.Timestamp,
		Metadata:   md,
		CreatedAt:  model.CreatedAt,
	}
}

func toRuleModel(rule *domain.Rule) *RuleModel {
	tags := make([]string, len(rule.Tags))
	copy(tags, rule.Tags)
	return &RuleModel{
		ID:          rule.ID,
		Name:        rule.Name,
		Pattern:     rule.Pattern,
		Description: rule.Description,
		Severity:    string(rule.Severity),
		Tags:        tags,
		Active:      rule.Active,
	}
}

func toDomainRule(model *RuleModel) *domain.Rule {
	if model == nil {
		return nil
	}
	tags := make([]string, len(model.Tags))
	copy(tags, model.Tags)
	return &domain.Rule{
		ID:          model.ID,
		Name:        model.Name,
		Pattern:     model.Pattern,
		Description: model.Description,
		Severity:    domain.SeverityLevel(model.Severity),
		Tags:        tags,
		Active:      model.Active,
		CreatedAt:   model.CreatedAt,
		UpdatedAt:   model.UpdatedAt,
	}
}

func toDetectionModel(detection *domain.Detection) *DetectionModel {
	metadata := make(map[string]any)
	if detection.Metadata != nil {
		for k, v := range detection.Metadata {
			metadata[k] = v
		}
	}
	matchedAt := detection.MatchedAt
	if matchedAt.IsZero() {
		matchedAt = time.Now().UTC()
	}
	return &DetectionModel{
		ID:        detection.ID,
		EventID:   detection.EventID,
		RuleID:    detection.RuleID,
		Summary:   detection.Summary,
		Severity:  string(detection.Severity),
		MatchedAt: matchedAt,
		Metadata:  metadata,
	}
}

func toDomainDetection(model *DetectionModel) *domain.Detection {
	if model == nil {
		return nil
	}
	metadata := make(map[string]any)
	for k, v := range model.Metadata {
		metadata[k] = v
	}
	return &domain.Detection{
		ID:        model.ID,
		EventID:   model.EventID,
		RuleID:    model.RuleID,
		Summary:   model.Summary,
		Severity:  domain.SeverityLevel(model.Severity),
		MatchedAt: model.MatchedAt,
		Metadata:  metadata,
	}
}

func toDecisionModel(decision *domain.Decision) *DecisionModel {
	return &DecisionModel{
		ID:          decision.ID,
		DetectionID: decision.DetectionID,
		Action:      string(decision.Action),
		Target:      decision.Target,
		Reason:      decision.Reason,
		Notes:       decision.Notes,
	}
}

func toDomainDecision(model *DecisionModel) *domain.Decision {
	if model == nil {
		return nil
	}
	return &domain.Decision{
		ID:          model.ID,
		DetectionID: model.DetectionID,
		Action:      domain.DecisionAction(model.Action),
		Target:      model.Target,
		Reason:      model.Reason,
		Notes:       model.Notes,
		CreatedAt:   model.CreatedAt,
	}
}
