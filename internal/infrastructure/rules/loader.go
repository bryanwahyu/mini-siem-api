package rules

import (
    "os"
    "strings"
    "time"
    "regexp"

    yaml "gopkg.in/yaml.v3"
    "server-analyst/internal/domain/services"
)

type RuleYAML struct {
    Name     string `yaml:"name"`
    Category string `yaml:"category"`
    Pattern  string `yaml:"pattern"`
    Enabled  bool   `yaml:"enabled"`
    Severity string `yaml:"severity"`
}

type FileRules struct {
    UpdatedAt time.Time  `yaml:"-"`
    Rules     []RuleYAML `yaml:"rules"`
}

func LoadRules(path string) (*FileRules, error) {
    data, err := os.ReadFile(path)
    if err != nil { return nil, err }
    var fr FileRules
    if err := yaml.Unmarshal(data, &fr); err != nil { return nil, err }
    fr.UpdatedAt = time.Now()
    return &fr, nil
}

func Compile(fr *FileRules) []services.CompiledRule {
    out := make([]services.CompiledRule, 0, len(fr.Rules))
    for _, r := range fr.Rules {
        var re *regexp.Regexp
        if strings.TrimSpace(r.Pattern) != "" {
            re = regexp.MustCompile(r.Pattern)
        }
        out = append(out, services.CompiledRule{Name: r.Name, Category: r.Category, Severity: r.Severity, Pattern: re, Enabled: r.Enabled})
    }
    return out
}
