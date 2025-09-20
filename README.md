# Server Analyst

Server Analyst is a threat analytics engine focused on catching common infrastructure attacks (JUDOL spam, brute force, SQL injection, XSS, traversal, scanner/fuzzer traffic, HTTP flood). The codebase follows Domain-Driven Design and ships as a single Go binary that can run as a systemd service or inside Docker, while archiving artefacts to MinIO with an offline-first spool.

## Features
- DDD boundaries (`internal/domain`, `internal/application`, `internal/infrastructure`, `internal/interfaces`).
- Multiple ingest paths (journald, tailing files, HTTP API) converging into a shared detection pipeline.
- Regex-based rule engine with per-rule metadata and enable/disable controls.
- Automated decisions that can trigger UFW/nftables/Cloudflare actions plus notifications.
- Persistence adapters for SQLite (default) and PostgreSQL via GORM.
- Cold storage uploads to MinIO with disk spooling when the bucket is unavailable.
- Structured logging with Zerolog, Prometheus metrics, Swagger-documented HTTP API, and CLI tooling.
- Automation scripts for provisioning, backups, firewall setup, systemd, and sample attack generation.

## Project Layout
- `cmd/server-analyst`: binary entrypoint.
- `internal/domain`: entities, value objects, services, repositories, specs.
- `internal/application`: DTOs, ports, and use cases orchestrating the domain.
- `internal/infrastructure`: adapters for persistence (SQLite/Postgres, MinIO), ingest tailers, notifications, rules/config loading, security actions, and observability.
- `internal/interfaces`: HTTP API (mux + swagger) and CLI entrypoints.
- `config`: runtime configuration (`config.yaml`, rules, allow/block lists, keyword maps).
- `scripts`: operational helpers (install, Docker/minio setup, firewall configuration, backups, sample generators).
- `docs`: generated Swagger assets. Regenerate with `swag init` (see Development section).

See `docs/ARCHITECTURE.md` for a deeper explanation of the components and message flow.

```
Ingest -> Detect -> Decide -> Act -> Archive
   |        |         |        |        \
 journal/  rules     policy  nft/ufw   MinIO
 files      +         +       nginx     (spool offline)
 parser   thresholds  allow/ blocklist  snapshots/backups
```

## Quick Start
1. Install Go 1.22+ and (optionally) Docker, MinIO CLI, and nftables/ufw if you plan to execute actions.
2. Build the binary: `make -C server-analyst build`
3. Run with the default configuration:
   ```bash
   ./server-analyst/server-analyst -config server-analyst/config/config.yaml
   ```
4. Verify the service:
   ```bash
   curl -s http://localhost:8080/health | jq .
   ```
5. Generate sample attacks and detections:
   ```bash
   server-analyst/scripts/generate_attack_samples.sh
   ```

### Docker Compose
```
cd server-analyst/docker
docker compose up --build
```
The stack launches the analyst service plus MinIO for cold storage testing.

## Configuration
- Base config: `config/config.yaml` (listener ports, DB DSN, actions, notifier settings).
- Detection rules: `config/rules.yml`, `config/keywords_judol.yml`.
- Policy lists: `config/allowlist.yml`, `config/blocklist.yml`.
- Cold storage: `MINIO_*` environment variables (see MinIO section).
- Override per-environment values by copying the YAML files and passing custom paths via CLI flags or env vars.

## Persistence and Archival
- Default storage is SQLite (`app.db`). Configure PostgreSQL via DSN in `config.yaml`.
- Cold storage uploads (detections snapshots, DB backups) go to MinIO/S3. If the bucket is down, payloads are spooled on disk and retried later.
- `scripts/backup_to_minio.sh` runs daily exports and SQLite backups.

## Actions and Notifications
- Firewall adapters exist for nftables, UFW, and nginx allow/deny management (`internal/infrastructure/actions`).
- Additional integrations include Cloudflare and Fail2ban. Enable them via config sections.
- Notifications (Slack/webhook/email) plug into `internal/infrastructure/notify`.

## Observability
- Structured JSON logs land in `logs/app.log` by default.
- Prometheus metrics are exposed at `GET /metrics`.
- Swagger UI is available at `http://<host>:<port>/swagger/index.html`.

## HTTP API
- Base path `/`, security via `ApiKeyAuth` header `X-API-Key` for admin endpoints.
- Endpoints: `GET /health`, `/events`, `/decisions`, `/detections`, `/metrics`, `POST /decisions/block`, `POST /decisions/unblock`.
- Example calls:
  ```bash
  curl -s http://localhost:8080/events | jq .
  curl -s "http://localhost:8080/detections?limit=100&offset=0" | jq .
  curl -s -X POST -H 'Content-Type: application/json' -H 'X-API-Key: <KEY>' \
       -d '{"ip":"1.2.3.4","reason":"manual block"}' \
       http://localhost:8080/decisions/block -i
  ```

## Installation (systemd)
Run on a fresh Ubuntu/Debian host as root:
```
server-analyst/scripts/install_dependencies.sh
server-analyst/scripts/install_app.sh
server-analyst/scripts/setup_systemd.sh
```
The service runs as the `srv-analyst` user with hardened systemd unit settings (`NoNewPrivileges`, protected filesystem, limited capabilities).

## MinIO Setup
```
export MINIO_ENDPOINT=minio.local:9000
export MINIO_ACCESS_KEY=MINIO_ACCESS_KEY
export MINIO_SECRET_KEY=MINIO_SECRET_KEY
export MINIO_BUCKET=server-analyst
server-analyst/scripts/minio_setup.sh
```
Spool data is retried automatically once the bucket is reachable.

## Development
- Run tests: `make -C server-analyst test`
- Format / tidy: `make -C server-analyst tidy`
- Regenerate Swagger docs:
  ```bash
  go install github.com/swaggo/swag/cmd/swag@v1.16.3
  cd server-analyst
  "$HOME/go/bin/swag" init -g cmd/server-analyst/main.go --parseDependency --parseInternal -o docs
  ```

## Security & Hardening
- Ship with `dry_run: true` until actions are verified in your environment.
- Lock down admin API by rotating the API key and restricting network access.
- Run the service as a non-privileged user; limit filesystem permissions to config and spool directories.
- Monitor Prometheus metrics and logs to tune thresholds and rule sets.

## Additional Documentation
- Architecture details and sequence diagrams live in `docs/ARCHITECTURE.md`.
- Configuration samples reside under `config/`.
