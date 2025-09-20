package main

// Server-Analyst API Swagger docs
// @title           Server-Analyst API
// @version         1.0
// @description     Threat detection & decision API
// @BasePath        /
// @schemes         http
// @securityDefinitions.apikey ApiKeyAuth
// @in              header
// @name            X-API-Key

import (
    "context"
    "flag"
    "fmt"
    "os"
    "os/signal"
    "regexp"
    "syscall"
    "time"

    "server-analyst/internal/application/usecases"
    "server-analyst/internal/domain/services"
    "server-analyst/internal/infrastructure/actions"
    cfg "server-analyst/internal/infrastructure/config"
    minioad "server-analyst/internal/infrastructure/persistence/minio"
    postgresad "server-analyst/internal/infrastructure/persistence/postgres"
    "server-analyst/internal/infrastructure/observability"
    "server-analyst/internal/infrastructure/notify"
    rloader "server-analyst/internal/infrastructure/rules"
    "server-analyst/internal/interfaces/cli"
    "server-analyst/internal/interfaces/httpapi"
)

func main() {
    cfgPath := flag.String("config", "", "config file path")
    flag.Parse()
    if err := observability.InitLogger("logs/app.log"); err != nil { fmt.Fprintln(os.Stderr, err) }

    conf, err := cfg.Load(*cfgPath)
    if err != nil { fmt.Fprintln(os.Stderr, "config error:", err); os.Exit(1) }

    // DB
    db, err := postgresad.Open(conf.Storage.DB.DSN)
    if err != nil { fmt.Fprintln(os.Stderr, "db error:", err); os.Exit(1) }
    eventRepo := postgresad.NewEventRepo(db.DB)
    decisionRepo := postgresad.NewDecisionRepo(db.DB)
    detectionSaver := postgresad.NewDetectionSaver(db.DB)
    spoolRepo := postgresad.NewSpoolRepo(db.DB)
    detectionRepo := postgresad.NewDetectionRepo(db.DB)

    // MinIO
    mc, err := minioad.New(conf.Storage.MinIO.Endpoint, conf.Storage.MinIO.UseSSL, conf.Storage.MinIO.AccessKey, conf.Storage.MinIO.SecretKey, conf.Storage.MinIO.Region, conf.Storage.MinIO.TimeoutSec, conf.Storage.MinIO.MaxRetries)
    if err != nil { fmt.Fprintln(os.Stderr, "minio error:", err) }
    cold := minioad.NewColdStore(mc, conf.Storage.MinIO.Bucket, conf.Storage.MinIO.Prefix, conf.Storage.MinIO.Region, conf.Storage.MinIO.SpoolDir)
    cold.SpoolMeta = spoolRepo
    ctx, cancel := context.WithCancel(context.Background())
    defer cancel()
    cold.StartBackgroundUploader(ctx)

    // Rules and detectors
    re := services.NewRuleEngine()
    judolKws := defaultJudolKeywords()
    // Load rules file if exists and watch
    snapUC := &usecases.SnapshotRulesUsecase{ColdStore: cold}
    rulesPath := conf.Rules.File
    if _, err := os.Stat(rulesPath); err == nil {
        if fr, err := rloader.LoadRules(rulesPath); err == nil { re.SetRules(rloader.Compile(fr)) }
    }
    // watch rules and keywords periodically
    go func() {
        ticker := time.NewTicker(15 * time.Second)
        defer ticker.Stop()
        rc := &rloader.Cache{}
        kc := &rloader.Cache{}
        for {
            select { case <-ctx.Done(): return; case <-ticker.C: }
            if rulesPath != "" {
                if changed, data, err := rc.LoadIfChanged(rulesPath); err == nil && changed {
                    if fr, err := rloader.LoadRules(rulesPath); err == nil { re.SetRules(rloader.Compile(fr)) }
                    _ = snapUC.Snapshot(ctx, data)
                }
            }
            if conf.Rules.JUDOLKeywords != "" {
                if changed, data, err := kc.LoadIfChanged(conf.Rules.JUDOLKeywords); err == nil && changed {
                    if kws, err := rloader.CompileKeywords(data); err == nil { judolKws = kws }
                }
            }
        }
    }()
    det := services.NewDetectorService(re, conf.Detect.Windows.Flood.Duration, conf.Detect.Windows.BruteForce.Duration, conf.Detect.Thresholds.SSHFailed, conf.Detect.Thresholds.HTTP401, conf.Detect.Thresholds.RPSPerIP, judolKws)
    dec := services.NewDecisionService()

    ingestUC := &usecases.IngestLogsUsecase{Events: eventRepo}
    // Notifiers
    var multi notify.Multi
    if conf.Notify.SlackWebhook != "" { multi.Notifiers = append(multi.Notifiers, &notify.Slack{Webhook: conf.Notify.SlackWebhook}) }
    if conf.Notify.TelegramBotToken != "" && conf.Notify.TelegramChatID != "" { multi.Notifiers = append(multi.Notifiers, &notify.Telegram{BotToken: conf.Notify.TelegramBotToken, ChatID: conf.Notify.TelegramChatID}) }

    detectUC := &usecases.DetectThreatsUsecase{Events: eventRepo, Decisions: decisionRepo, Detector: det, Decider: dec, SaveDetection: detectionSaver.Save, Notifier: multi.Notify}
    exportUC := &usecases.ExportEventsUsecase{Events: eventRepo, ColdStore: cold}
    exportDetUC := &usecases.ExportDetectionsUsecase{Detections: detectionRepo, ColdStore: cold}

    // If args beyond flags exist, run CLI
    rest := flag.Args()
    if len(rest) > 0 {
        os.Exit(cli.Run(rest, cli.Deps{ExportUC: exportUC, ExportDetUC: exportDetUC, IngestUC: ingestUC, DetectUC: detectUC}))
        return
    }

    // Server
    // Build actions chain
    var act actions.Multi
    act.Ports = append(act.Ports, actions.NewAdapter(conf.Server.DryRun))
    if conf.Actions.Cloudflare.Enabled {
        act.Ports = append(act.Ports, &actions.Cloudflare{Enabled: true, APIToken: conf.Actions.Cloudflare.APIToken, ZoneID: conf.Actions.Cloudflare.ZoneID, DryRun: conf.Server.DryRun})
    }

    srv := httpapi.NewServer(httpapi.Deps{
        Cfg:            conf,
        ListEvents:     &usecases.ListEventsQuery{Repo: eventRepo},
        ListDecisions:  &usecases.ListDecisionsQuery{Repo: decisionRepo},
        ListDetections: &usecases.ListDetectionsQuery{Repo: detectionRepo},
        Action:         &usecases.ApplyActionUsecase{Port: &act},
    })
    if err := srv.Start(); err != nil { fmt.Fprintln(os.Stderr, "http error:", err); os.Exit(1) }
    observability.Logger.Info().Str("addr", conf.Server.HTTPAddr).Msg("server started")

    // actions adapter already wired for admin endpoints

    sig := make(chan os.Signal, 1)
    signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
    <-sig
    ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel2()
    _ = srv.Stop(ctx2)
}

// compile ~50+ JUDOL keywords into regexes (simple word match)
func defaultJudolKeywords() []*regexp.Regexp {
    list := []string{"slot", "judol", "gacor", "maxwin", "scatter", "rtp", "bonanza", "wild", "jackpot", "casino", "bet", "togel", "slot88", "slot77", "slot69", "pragmatic", "olxtoto", "gates of olympus", "habanero", "pgsoft", "microgaming", "betwin", "idr", "deposit", "withdraw", "garansi", "spin", "naga", "maxbet", "sbobet", "royal", "bonus new member", "bonus harian", "slotmania", "betting", "slotmania", "slotku", "slotindo", "slotviral", "slotmax", "slotking", "slotserver", "slotgila", "slotjitu", "slotmantap", "slotpulsa", "slotpulsa tanpa potongan", "gates", "olympus", "zeus", "sweet bonanza", "starlight", "princess", "mahjong", "ways", "riches", "aztec", "gems"}
    out := make([]*regexp.Regexp, 0, len(list))
    for _, w := range list {
        out = append(out, regexp.MustCompile("(?i)"+regexp.QuoteMeta(w)))
    }
    return out
}
