package usecases

import (
    "context"
    "strings"
    "time"

    "server-analyst/internal/domain/entities"
    "server-analyst/internal/domain/repositories"
    "server-analyst/internal/infrastructure/observability"
)

type IngestLogsUsecase struct {
    Events repositories.EventRepository
}

type LogInput struct {
    Host   string
    Source string
    Line   string
}

func (uc *IngestLogsUsecase) Ingest(ctx context.Context, in LogInput) (*entities.Event, error) {
    ev := &entities.Event{
        CreatedAt: time.Now(),
        Host:      in.Host,
        Source:    in.Source,
        Raw:       in.Line,
    }
    // naive parse for IP, method, path, status
    low := strings.ToLower(in.Line)
    for _, token := range strings.Split(low, " ") {
        if strings.Count(token, ".") == 3 && !strings.Contains(token, ":") {
            ev.IP = token
            break
        }
    }
    if i := strings.Index(in.Line, "\""); i >= 0 {
        // try to extract method and path in quotes
        rest := in.Line[i+1:]
        if j := strings.Index(rest, "\""); j >= 0 {
            req := rest[:j]
            parts := strings.Split(req, " ")
            if len(parts) >= 2 { ev.Method = parts[0]; ev.Path = parts[1] }
        }
    }
    if err := uc.Events.Save(ctx, ev); err != nil { return nil, err }
    observability.EventsTotal.Inc()
    return ev, nil
}
