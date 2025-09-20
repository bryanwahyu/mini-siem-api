package services

import (
    "regexp"
    "strings"
    "sync"
    "time"

    "server-analyst/internal/domain/entities"
)

type DetectorService struct {
    re *RuleEngine
    // sliding counters per IP
    mu           sync.Mutex
    failuresSSH  map[string][]time.Time
    http401      map[string][]time.Time
    rpsByIP      map[string][]time.Time
    floodWindow  time.Duration
    bruteWindow  time.Duration
    thrSSH       int
    thrHTTP401   int
    thrRPSPerIP  int
    judolKW      []*regexp.Regexp
}

func NewDetectorService(re *RuleEngine, flood, brute time.Duration, thrSSH, thr401, thrRPS int, judol []*regexp.Regexp) *DetectorService {
    return &DetectorService{
        re:          re,
        failuresSSH: make(map[string][]time.Time),
        http401:     make(map[string][]time.Time),
        rpsByIP:     make(map[string][]time.Time),
        floodWindow: flood,
        bruteWindow: brute,
        thrSSH:      thrSSH,
        thrHTTP401:  thr401,
        thrRPSPerIP: thrRPS,
        judolKW:     judol,
    }
}

func (d *DetectorService) Detect(ev *entities.Event) (out []entities.Detection) {
    // Keyword/regex rules
    for _, cr := range d.re.Rules() {
        if !cr.Enabled { continue }
        text := strings.Join([]string{ev.Raw, ev.Path, ev.Referrer, ev.UA, ev.Method}, "\n")
        if cr.Pattern != nil && cr.Pattern.MatchString(text) {
            out = append(out, entities.Detection{EventID: ev.ID, Category: cr.Category, Rule: cr.Name, Severity: cr.Severity})
        }
    }
    // JUDOL keywords
    if ev.Referrer != "" || ev.Path != "" || ev.Raw != "" {
        text := strings.ToLower(strings.Join([]string{ev.Referrer, ev.Path, ev.Raw}, "\n"))
        for _, kw := range d.judolKW {
            if kw.MatchString(text) {
                out = append(out, entities.Detection{EventID: ev.ID, Category: "judol", Rule: "keyword_match", Severity: "low"})
                break
            }
        }
    }
    // Sliding window counters for brute force and flood
    d.mu.Lock()
    defer d.mu.Unlock()
    now := time.Now()
    if ev.Source == "sshd" && strings.Contains(strings.ToLower(ev.Raw), "failed password") && ev.IP != "" {
        d.failuresSSH[ev.IP] = appendAndTrim(d.failuresSSH[ev.IP], now, d.bruteWindow)
        if len(d.failuresSSH[ev.IP]) >= d.thrSSH {
            out = append(out, entities.Detection{EventID: ev.ID, Category: "brute", Rule: "ssh_failed", Severity: "medium"})
        }
    }
    if ev.Status == 401 && ev.IP != "" {
        d.http401[ev.IP] = appendAndTrim(d.http401[ev.IP], now, d.bruteWindow)
        if len(d.http401[ev.IP]) >= d.thrHTTP401 {
            out = append(out, entities.Detection{EventID: ev.ID, Category: "brute", Rule: "http_401", Severity: "medium"})
        }
    }
    if ev.IP != "" && ev.Method != "" {
        d.rpsByIP[ev.IP] = appendAndTrim(d.rpsByIP[ev.IP], now, d.floodWindow)
        if len(d.rpsByIP[ev.IP]) >= d.thrRPSPerIP {
            out = append(out, entities.Detection{EventID: ev.ID, Category: "flood", Rule: "rps_per_ip", Severity: "high"})
        }
    }
    // Common payload signatures
    payload := strings.ToLower(strings.Join([]string{ev.Path, ev.Raw}, "\n"))
    if reSQLI.MatchString(payload) {
        out = append(out, entities.Detection{EventID: ev.ID, Category: "sqli", Rule: "sqli_regex", Severity: "high"})
    }
    if reXSS.MatchString(payload) {
        out = append(out, entities.Detection{EventID: ev.ID, Category: "xss", Rule: "xss_regex", Severity: "medium"})
    }
    if reTraversal.MatchString(payload) {
        out = append(out, entities.Detection{EventID: ev.ID, Category: "traversal", Rule: "path_traversal", Severity: "high"})
    }
    if reScanner.MatchString(payload) {
        out = append(out, entities.Detection{EventID: ev.ID, Category: "scanner", Rule: "scanner_signature", Severity: "low"})
    }
    return out
}

var (
    reSQLI      = regexp.MustCompile(`(?i)(union\s+select|select\s+.+\s+from|information_schema|or\s+1=1|\bupdate\b\s+.*set|insert\s+into|sleep\s*\(|benchmark\s*\()`)
    reXSS       = regexp.MustCompile(`(?i)(<script|onerror=|onload=|alert\(|document\.cookie|<img\s+src=|svg\s+onload=)`)
    reTraversal = regexp.MustCompile(`(?i)(\.\./|%2e%2e/|etc/passwd|proc/self/environ|php://input|file://)`)
    reScanner   = regexp.MustCompile(`(?i)(wp-admin|phpmyadmin|boaform|HNAP1|\bmanager\b|hudson|jenkins|adminer)`)
)

func appendAndTrim(ts []time.Time, now time.Time, window time.Duration) []time.Time {
    ts = append(ts, now)
    cutoff := now.Add(-window)
    i := 0
    for ; i < len(ts); i++ {
        if ts[i].After(cutoff) {
            break
        }
    }
    return ts[i:]
}
