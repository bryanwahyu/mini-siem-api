package cli

import (
    "bufio"
    "context"
    "flag"
    "fmt"
    "os"
    "strings"
    "server-analyst/internal/application/usecases"
)

func replayCmd(args []string, d Deps) int {
    fs := flag.NewFlagSet("replay", flag.ExitOnError)
    file := fs.String("file", "", "log file to replay")
    source := fs.String("source", "replay", "source tag")
    host := fs.String("host", hostname(), "host tag")
    fs.Parse(args)
    if *file == "" { fmt.Fprintln(os.Stderr, "-file required"); return 1 }
    f, err := os.Open(*file)
    if err != nil { fmt.Fprintln(os.Stderr, err); return 1 }
    defer f.Close()
    sc := bufio.NewScanner(f)
    ctx := context.Background()
    count := 0
    for sc.Scan() {
        line := strings.TrimRight(sc.Text(), "\n")
        ev, err := d.IngestUC.Ingest(ctx, usecases.LogInput{Host: *host, Source: *source, Line: line})
        if err != nil { fmt.Fprintln(os.Stderr, err); return 1 }
        _, _, _ = d.DetectUC.Process(ctx, ev)
        count++
    }
    fmt.Printf("replayed %d lines\n", count)
    return 0
}
