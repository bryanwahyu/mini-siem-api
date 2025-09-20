package usecases

import (
    "bufio"
    "bytes"
    "compress/gzip"
    "context"
    "encoding/json"
    "fmt"
    "time"
    "server-analyst/internal/domain/repositories"
)

type ExportEventsUsecase struct {
    Events    repositories.EventRepository
    ColdStore repositories.ColdStoreRepository
    Hostname  string
    Prefix    string
}

func (uc *ExportEventsUsecase) ExportDaily(ctx context.Context, day time.Time) error {
    // Export all events as NDJSON gz
    evs, err := uc.Events.List(ctx, 10000, 0)
    if err != nil { return err }
    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)
    w := bufio.NewWriter(gz)
    for _, ev := range evs {
        b, _ := json.Marshal(ev)
        w.Write(b)
        w.WriteByte('\n')
    }
    w.Flush(); gz.Close()
    object := fmt.Sprintf("%s/events/%04d/%02d/%02d/events-%s-%d.ndjson.gz", uc.Prefix, day.Year(), day.Month(), day.Day(), uc.Hostname, time.Now().Unix())
    if err := uc.ColdStore.Upload(ctx, object, "application/x-ndjson", true, buf.Bytes()); err != nil { return err }
    return nil
}
