package cli

import (
	"context"
	"flag"
	"fmt"
	"os"
	"time"

	"server-analyst/internal/application/usecases"
)

type Deps struct {
	ExportUC    *usecases.ExportEventsUsecase
	ExportDetUC *usecases.ExportDetectionsUsecase
	IngestUC    *usecases.IngestLogsUsecase
	DetectUC    *usecases.DetectThreatsUsecase
}

func Run(args []string, d Deps) int {
	if len(args) == 0 {
		fmt.Println("usage: server-analyst [serve|export|replay|health]")
		return 1
	}
	switch args[0] {
	case "export":
		fs := flag.NewFlagSet("export", flag.ExitOnError)
		host := fs.String("host", hostname(), "hostname tag")
		prefix := fs.String("prefix", "prod", "S3 prefix")
		typ := fs.String("type", "events", "what to export: events|detections")
		fs.Parse(args[1:])
		switch *typ {
		case "events":
			d.ExportUC.Hostname = *host
			d.ExportUC.Prefix = *prefix
			if err := d.ExportUC.ExportDaily(context.Background(), time.Now()); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		case "detections":
			d.ExportDetUC.Prefix = *prefix
			if err := d.ExportDetUC.ExportDaily(context.Background(), time.Now()); err != nil {
				fmt.Fprintln(os.Stderr, err)
				return 1
			}
		default:
			fmt.Fprintln(os.Stderr, "unsupported type")
			return 1
		}
		fmt.Println("exported")
		return 0
	case "replay":
		return replayCmd(args[1:], d)
	case "health":
		return healthCmd(args[1:])
	default:
		fmt.Println("unknown command:", args[0])
		return 1
	}
}

func hostname() string { h, _ := os.Hostname(); return h }
