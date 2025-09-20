# Architecture

Server Analyst follows a layered (DDD-inspired) structure. Core business logic is isolated in the domain layer, while adapters live in infrastructure and are orchestrated through application layer use cases. Interfaces (HTTP/CLI) depend on the application layer only.

```
                 ┌────────────────────────────────────────┐
                 │               Interfaces               │
                 │   HTTP API (mux)        CLI Commands   │
                 └────────────────────────────────────────┘
                                   ↓
                 ┌────────────────────────────────────────┐
                 │             Application Layer          │
                 │ Usecases (Detect, Ingest, Export, ...) │
                 │ DTOs / Ports / Queries                 │
                 └────────────────────────────────────────┘
                                   ↓
                 ┌────────────────────────────────────────┐
                 │               Domain Layer             │
                 │ Entities · Value Objects · Services    │
                 │ Repositories (interfaces) · Specs      │
                 └────────────────────────────────────────┘
                                   ↓
                 ┌────────────────────────────────────────┐
                 │           Infrastructure Layer         │
                 │ Persistence · Actions · Tailers ·      │
                 │ Notifications · Config · Spool · Rules │
                 └────────────────────────────────────────┘
```

## Core Flow

1. **Ingest** tailers (journald/file) or API feed raw log lines into the `IngestLogsUsecase`.
2. Events are persisted through repository interfaces (SQLite/Postgres adapters) while simultaneously passed to the detection pipeline.
3. `DetectorService` runs signature/regex rules, keyword checks, and sliding window counters (brute force, HTTP 401 bursts, rate limiting).
4. `DecisionService` applies simple policies (e.g., block IP for 1h on brute/flood) and persists decisions.
5. Application layer triggers infrastructure actions (firewall adapters, notifications) and updates Prometheus counters.
6. Artefacts (events, detections, DB snapshots, rule snapshots) are streamed to MinIO cold storage. When the bucket is unreachable, the payload is stored on disk (`spool/`) and a retry job flushes it later.

```
Logs (journald/files/API)
        │
        ▼
Ingest Usecase ──► Event Repository (SQLite/Postgres)
        │
        ▼
DetectorService (rule engine + heuristics)
        │             ┌───────────────┐
        │             │Notifications  │
        ├────────────►│(slack/webhook)│
        │             └───────────────┘
        ▼
DecisionService ──► Decision Repository
        │
        ├────────────► Actions (UFW/nftables/Cloudflare)
        ▼
Cold Store (MinIO) ◄─ Disk Spool (retry worker)
```

## Domain Layer (`internal/domain`)
- **Entities**: `Event`, `Detection`, `Decision`, `SpoolItem` capture the persisted state.
- **Services**:
  - `RuleEngine` stores compiled regex rules updated from YAML files.
  - `DetectorService` applies rule patterns, keyword lists, and sliding-window thresholds to incoming events.
  - `DecisionService` maps detection categories to enforcement actions (currently block-on-brute/flood with TTL).
- **Repositories** define behaviour for persistence without binding to a specific database.
- **Specs/Value Objects**: enforce business rules and shared value semantics (e.g., IP, status validation).

## Application Layer (`internal/application`)
- **Usecases** coordinate domain services and repositories: `DetectThreatsUsecase`, `IngestLogsUsecase`, `ApplyActionUsecase`, `SnapshotRulesUsecase`, export jobs, queries for listing events/detections/decisions.
- **DTOs** adapt domain entities for transport over HTTP/CLI.
- **Ports** describe external capabilities (`Tailer`, `Notifier`) that infrastructure implements.

## Infrastructure Layer (`internal/infrastructure`)
- **Persistence**:
  - `persistence/sqlite` and `persistence/postgres` expose GORM-backed repositories for events/detections/decisions/spool metadata.
  - `persistence/minio` handles cold storage uploads, rule snapshots, and background spool flushing with retry/backoff.
- **Spool**: `spool.DiskSpool` writes payloads to disk with metadata tracked in the database to guarantee eventual upload to MinIO.
- **Log sources**: file tailer (polling) and journald tailer (stubbed in minimal build, Linux-only implementation can be plugged in).
- **Actions**: adapters for UFW, nftables, nginx ACLs, Fail2ban, and Cloudflare. `actions.Multi` fans out to multiple backends.
- **Notifications**: pluggable notifiers for Slack/webhook/email.
- **Rules**: loaders that parse YAML rule files and feed `RuleEngine`.
- **Config**: typed configuration loading with defaults/validation.
- **Security**: API key middleware and simple rate limiting for the admin routes.
- **Observability**: Prometheus collectors and Zerolog setup for structured logging.

## Interfaces Layer (`internal/interfaces`)
- **HTTP API** (`httpapi`): gorilla/mux router, Swagger docs, Prometheus metrics endpoint, admin routes secured by API key + rate limiter.
- **CLI** (`cli`): commands for exports, replaying sample logs, and health checks; composes use cases for offline automation.

## Persistence Model
- **Operational DB**: SQLite by default (`app.db`), switchable to PostgreSQL via DSN. Repositories persist events, detections, decisions, and spool metadata.
- **Cold Storage**: MinIO/S3 bucket receives NDJSON or compressed archives. Uploads retry with exponential backoff. Failed uploads drop into the disk spool for later replay (`FlushSpool`). `SnapshotRulesUsecase` stores timestamped rule dumps under `rules/snapshots/`.

## Resilience & Back Pressure
- Spool metadata keeps track of retries and last errors to aid observability.
- Background uploader drains spool every 30s; manual triggers are available via CLI export commands.
- Actions operate in dry-run mode until explicitly enabled to avoid accidental enforcement.
- Detection thresholds (rate limits, brute force counters) are configurable via YAML allowing tuning per environment.

## Observability & Instrumentation
- Prometheus counters: events processed, detections per category, decisions per action, spool queue size, upload successes/failures.
- Logs: Zerolog JSON logs aimed at `logs/app.log`, integrate with journald or file tailing.
- Health check: `/health` for liveness; `/metrics` for exporters; Swagger UI for API discovery.

## Deployment Considerations
- **Systemd**: install scripts create a dedicated `srv-analyst` user, apply `NoNewPrivileges`, `ProtectSystem`, and capability drops.
- **Docker**: `docker/docker-compose.yml` builds the binary and runs alongside MinIO for development/testing.
- **Configuration Management**: environment vars override YAML; secrets (API keys, MinIO creds) should be injected via `.env` or secret stores.
- **Scaling**: run multiple instances behind a queue or log forwarder; detectors are stateless apart from in-memory sliding windows.

## Future Extensions
- Plug in richer detection sources (Netflow, Zeek) by implementing the `Tailer` or pushing events via HTTP.
- Add dedicated detection repository separate from events for analytics-heavy workloads.
- Extend decision service for allow-list decay, automated unblock policies, and integration with SOAR platforms.

## Mini SIEM API (Standalone Service)

The `mini-siem-api/` directory contains a trimmed-down HTTP service derived from the Server Analyst concepts. It follows the same domain/application/infra layering and exposes a Swagger-documented REST API for:

- Event ingestion with rate limiting and batch support
- Rule management (create/list/toggle) guarded by API key middleware
- Detection listing and detail views with time/severity filters
- Analyst decisions audit trail
- Observability endpoints (`/health`, `/metrics`, `/swagger`)

Events, detections, and decisions are persisted through GORM repositories (SQLite or PostgreSQL) with optional cold storage uploads via MinIO. The detection engine reuses the regex rule engine and is warmed on startup. Docker assets (`mini-siem-api/Dockerfile`, `mini-siem-api/docker-compose.yaml`) allow running the API alongside PostgreSQL for local testing.
