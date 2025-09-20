package services

import (
    "time"
    "server-analyst/internal/domain/entities"
)

type DecisionService struct{}

func NewDecisionService() *DecisionService { return &DecisionService{} }

// Decide simple policy: if category brute or flood -> block for 1h
func (d *DecisionService) Decide(dets []entities.Detection, ip string) *entities.Decision {
    if ip == "" || len(dets) == 0 { return nil }
    var cat string
    for _, de := range dets {
        if de.Category == "brute" || de.Category == "flood" {
            cat = de.Category
            break
        }
    }
    if cat == "" { return nil }
    until := time.Now().Add(1 * time.Hour)
    return &entities.Decision{IP: ip, Action: "block", Reason: cat, Until: &until}
}
