package services

import (
    "server-analyst/internal/domain/entities"
    "testing"
    "time"
)

func TestSQLIDetection(t *testing.T) {
    re := NewRuleEngine()
    det := NewDetectorService(re, time.Minute, time.Minute, 8, 20, 120, nil)
    ev := &entities.Event{Raw: "GET /?q=union select 1 HTTP/1.1"}
    ds := det.Detect(ev)
    found := false
    for _, d := range ds { if d.Category=="sqli" { found = true; break } }
    if !found { t.Fatalf("expected sqli detection") }
}

func TestFloodThreshold(t *testing.T) {
    re := NewRuleEngine()
    det := NewDetectorService(re, time.Second, time.Minute, 8, 20, 10, nil)
    ip := "198.51.100.23"
    for i:=0;i<12;i++ {
        ev := &entities.Event{IP: ip, Method: "GET", Path: "/"}
        ds := det.Detect(ev)
        if i < 9 && len(ds)>0 { t.Fatalf("unexpected detection before threshold") }
    }
    ev := &entities.Event{IP: ip, Method: "GET", Path: "/"}
    ds := det.Detect(ev)
    ok := false
    for _, d := range ds { if d.Category=="flood" { ok=true } }
    if !ok { t.Fatalf("expected flood detection after threshold") }
}
