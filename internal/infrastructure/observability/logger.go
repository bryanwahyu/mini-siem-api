package observability

import (
    "io"
    stdlog "log"
    "os"
    "path/filepath"
    "time"

    "github.com/rs/zerolog"
    zlog "github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func InitLogger(logFilePath string) error {
    // Ensure dir exists
    if logFilePath != "" {
        if err := os.MkdirAll(filepath.Dir(logFilePath), 0o750); err != nil {
            return err
        }
    }
    var writers []io.Writer
    // JSON to stdout
    writers = append(writers, os.Stdout)

    if logFilePath != "" {
        f, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
        if err == nil {
            writers = append(writers, f)
        }
    }

    mw := io.MultiWriter(writers...)
    zerolog.TimeFieldFormat = time.RFC3339
    l := zerolog.New(mw).With().Timestamp().Logger()
    zlog.Logger = l
    Logger = l
    // Redirect stdlib
    std := zlog.Logger
    stdlog.SetFlags(0)
    stdlog.SetOutput(std)
    return nil
}
