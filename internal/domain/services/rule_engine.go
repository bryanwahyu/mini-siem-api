package services

import (
    "regexp"
    "sync"
)

type CompiledRule struct {
    Name     string
    Category string
    Severity string
    Pattern  *regexp.Regexp
    Enabled  bool
}

type RuleEngine struct {
    mu    sync.RWMutex
    rules []CompiledRule
}

func NewRuleEngine() *RuleEngine { return &RuleEngine{} }

func (re *RuleEngine) SetRules(rules []CompiledRule) {
    re.mu.Lock()
    defer re.mu.Unlock()
    re.rules = rules
}

func (re *RuleEngine) Rules() []CompiledRule {
    re.mu.RLock()
    defer re.mu.RUnlock()
    out := make([]CompiledRule, len(re.rules))
    copy(out, re.rules)
    return out
}
