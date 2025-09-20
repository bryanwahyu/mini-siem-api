package rules

import (
    "regexp"
    "strings"

    yaml "gopkg.in/yaml.v3"
)

type KeywordsFile struct {
    Keywords []string `yaml:"keywords"`
}

func CompileKeywords(data []byte) ([]*regexp.Regexp, error) {
    var kf KeywordsFile
    if err := yaml.Unmarshal(data, &kf); err != nil { return nil, err }
    out := make([]*regexp.Regexp, 0, len(kf.Keywords))
    for _, w := range kf.Keywords {
        w = strings.TrimSpace(w)
        if w == "" { continue }
        out = append(out, regexp.MustCompile("(?i)"+regexp.QuoteMeta(w)))
    }
    return out, nil
}

