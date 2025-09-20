package usecases

import (
    "context"
    "encoding/json"

    "server-analyst/internal/domain/entities"
    "server-analyst/internal/domain/repositories"
    "server-analyst/internal/domain/services"
    "server-analyst/internal/infrastructure/observability"
)

type DetectThreatsUsecase struct {
    Detections repositories.EventRepository // not ideal; we'll add a DetectionRepository via EventRepository adapter methods
    Events     repositories.EventRepository
    Decisions  repositories.DecisionRepository
    Detector   *services.DetectorService
    Decider    *services.DecisionService
    SaveDetection func(ctx context.Context, det *entities.Detection) error
    Notifier func(ctx context.Context, title, text string) error
}

func (uc *DetectThreatsUsecase) Process(ctx context.Context, ev *entities.Event) ([]entities.Detection, *entities.Decision, error) {
    dets := uc.Detector.Detect(ev)
    for i := range dets {
        d := &dets[i]
        if d.Metadata == "" {
            b, _ := json.Marshal(map[string]any{"ip": ev.IP, "path": ev.Path, "host": ev.Host, "source": ev.Source})
            d.Metadata = string(b)
        }
        if err := uc.SaveDetection(ctx, d); err != nil { return nil, nil, err }
        observability.DetectionsTotal.WithLabelValues(d.Category, d.Rule).Inc()
    }
    if len(dets) > 0 && uc.Notifier != nil {
        _ = uc.Notifier(ctx, "Detections", string(mustJSON(map[string]any{"ip": ev.IP, "count": len(dets), "path": ev.Path})))
    }
    decision := uc.Decider.Decide(dets, ev.IP)
    if decision != nil {
        if err := uc.Decisions.Save(ctx, decision); err != nil { return dets, nil, err }
        observability.DecisionsTotal.WithLabelValues(decision.Action).Inc()
    }
    return dets, decision, nil
}

func mustJSON(v any) []byte { b, _ := json.Marshal(v); return b }
