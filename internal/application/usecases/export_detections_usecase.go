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

type ExportDetectionsUsecase struct {
    Detections repositories.DetectionRepository
    ColdStore  repositories.ColdStoreRepository
    Prefix     string
}

func (uc *ExportDetectionsUsecase) ExportDaily(ctx context.Context, day time.Time) error {
    dets, err := uc.Detections.List(ctx, 10000, 0)
    if err != nil { return err }
    var buf bytes.Buffer
    gz := gzip.NewWriter(&buf)
    w := bufio.NewWriter(gz)
    for _, d := range dets {
        b, _ := json.Marshal(d)
        w.Write(b)
        w.WriteByte('\n')
    }
    w.Flush(); gz.Close()
    object := fmt.Sprintf("%s/detections/%04d/%02d/%02d/detections-%d.ndjson.gz", uc.Prefix, day.Year(), day.Month(), day.Day(), time.Now().Unix())
    return uc.ColdStore.Upload(ctx, object, "application/x-ndjson", true, buf.Bytes())
}

