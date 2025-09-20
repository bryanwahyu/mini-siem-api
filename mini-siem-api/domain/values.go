package domain

import (
	"fmt"
	"strings"
)

// ParseSeverity converts an external string to a SeverityLevel.
func ParseSeverity(input string) (SeverityLevel, error) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	switch normalized {
	case string(SeverityLow), "":
		return SeverityLow, nil
	case string(SeverityMedium):
		return SeverityMedium, nil
	case string(SeverityHigh):
		return SeverityHigh, nil
	case string(SeverityCritical):
		return SeverityCritical, nil
	default:
		return SeverityLow, fmt.Errorf("unknown severity '%s'", input)
	}
}

// ParseDecisionAction validates the requested action value.
func ParseDecisionAction(input string) (DecisionAction, error) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	switch normalized {
	case string(DecisionBlock):
		return DecisionBlock, nil
	case string(DecisionIgnore):
		return DecisionIgnore, nil
	case string(DecisionMonitor):
		return DecisionMonitor, nil
	default:
		return DecisionIgnore, fmt.Errorf("unknown decision action '%s'", input)
	}
}
