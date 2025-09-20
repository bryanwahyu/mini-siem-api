# Mini SIEM API

Mini SIEM API is a lightweight event ingestion and anomaly detection service exposing a REST API with Swagger documentation. It stores events and detections in PostgreSQL (SQLite supported for local development) and keeps a simple rule engine for regex-based detections.

## Features
- Event ingestion (`POST /events`) with batch support and rate limiting
- Rule-based detections with enable/disable controls
- Detection and decision retrieval endpoints
- API key protection for administrative routes (rules, decisions)
- Prometheus metrics (`GET /metrics`) and health check (`GET /health`)
- Generated Swagger docs served at `/swagger/index.html`

## Getting Started

### Prerequisites
- Go 1.22+
- PostgreSQL (or SQLite for quick local testing)

### Environment Variables
See `.env.example` for common variables. The service reads configuration purely from environment variables (`MINI_SIEM_*`). Key settings:

- `MINI_SIEM_ADDRESS`: HTTP listen address (default `:8085`)
- `MINI_SIEM_ADMIN_API_KEY`: shared secret for admin endpoints
- `MINI_SIEM_DB_DRIVER`: `postgres` or `sqlite`
- `MINI_SIEM_DB_DSN`: database DSN/URL
- `MINI_SIEM_INGEST_RATE`, `MINI_SIEM_INGEST_BURST`: rate limiter settings for the ingest endpoint

### Local Run
```bash
cd mini-siem-api
cp .env.example .env  # adjust as needed
export $(grep -v '^#' .env | xargs)
go run ./cmd/mini-siem-api
```

Once running:
- Health check: `curl -s http://localhost:8085/health | jq .`
- Swagger UI: `http://localhost:8085/swagger/index.html`

### Docker Compose
A compose file is provided to run the API with PostgreSQL:
```bash
cd mini-siem-api
docker compose up --build
```
The API listens on `http://localhost:8085` and the database on `localhost:5433`.

## Example Requests

Ingest a single event:
```bash
curl -s -X POST http://localhost:8085/events \
  -H 'Content-Type: application/json' \
  -d '{"source":"web","ip":"203.0.113.1","message":"failed login from 203.0.113.1","metadata":{"user":"alice"}}'
```

List stored events with optional filters:
```bash
curl -s "http://localhost:8085/events?source=web&limit=10" | jq .
```

Fetch a specific event:
```bash
curl -s http://localhost:8085/events/1 | jq .
```

List detections:
```bash
curl -s "http://localhost:8085/detections?severity=high" | jq .
```

Create a rule (requires admin API key):
```bash
curl -s -X POST http://localhost:8085/rules \
  -H 'Content-Type: application/json' \
  -H 'X-API-Key: changeme' \
  -d '{"name":"SSH brute","pattern":"failed login","severity":"high"}' | jq .
```

## Testing
Run `go test ./...` from `mini-siem-api` to execute package tests (none provided by default). `go build ./...` verifies compilation.

## Swagger Docs
Swagger artifacts are generated under `api/docs`. Regenerate after handler changes:
```bash
swag init -g cmd/mini-siem-api/main.go --parseDependency --parseInternal -o api/docs
```

## Project Layout
- `cmd/mini-siem-api`: application entrypoint
- `api/`: HTTP handlers, middleware, Swagger definitions
- `config/`: configuration helpers
- `domain/`: core entities and interfaces
- `infra/`: persistence, services, rate limiters, storage adapters

