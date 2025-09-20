package minio

import (
    "context"
    "os"
    "path/filepath"
    "testing"
)

func TestUploadSpoolsWhenMinIODown(t *testing.T) {
    c, _ := New("127.0.0.1:65535", false, "a", "b", "us-east-1", 1, 1)
    dir := t.TempDir()
    cs := NewColdStore(c, "bucket", "prefix", "us-east-1", dir)
    data := []byte("{"+"\"k\":1}")
    if err := cs.Upload(context.Background(), "events/2024/09/10/events-host-1.ndjson.gz", "application/x-ndjson", true, data); err != nil {
        // Upload returns error after spooling
    }
    files, err := os.ReadDir(dir)
    if err != nil { t.Fatal(err) }
    if len(files) == 0 { t.Fatalf("expected spooled file in %s", dir) }
    if filepath.Ext(files[0].Name()) != ".part" { t.Fatalf("expected .part file") }
}

