# Rig-to-Cloud Telemetry Monorepo

Polyglot monorepo for a resilient rig-to-cloud drilling telemetry pipeline with WITS/serial emulation.

> **Dokumentasi lengkap (Bahasa Indonesia):** [`docs/PROJECT-AS-IS.md`](docs/PROJECT-AS-IS.md)

## Layout

| Path | Role |
|------|------|
| `edge-daemon/` | Go edge agent — serial/WITS ingest, SQLite outbox, MQTT sync |
| `cloud-backend/` | Go cloud ingest — MQTT → TimescaleDB, HTTP/WebSocket API |
| `shared/` | Shared Go models + MQTT topic helpers (`rigtelemetry/shared`) |
| `simulator/` | Python WITS generator (PTY/serial writer) |
| `dashboard/` | React live well-log dashboard |
| `infrastructure/` | Mosquitto & TimescaleDB init configs |

## Prerequisites

- Go 1.22+ (CGO/gcc for `go-sqlite3`)
- Python 3.10+
- Node.js 20+ / npm
- Docker & Docker Compose

## Quick start (end-to-end)

```bash
make init
make infra-up          # TimescaleDB :5433, Mosquitto :1883/:9001

# Terminal A — cloud API + MQTT consumer
make run-cloud

# Terminal B — WITS simulator (prints SERIAL_PORT=...)
make run-sim

# Terminal C — edge daemon (use the printed PTY path)
SERIAL_PORT=/dev/pts/N make run-edge

# Terminal D — dashboard
make run-dashboard     # http://localhost:5173
```

Copy env defaults from [`.env.example`](.env.example) as needed.

## Makefile targets

| Target | Description |
|--------|-------------|
| `make init` | Go tidy + Python venv + npm install |
| `make infra-up` / `infra-down` | Start/stop Docker services |
| `make run-sim` | WITS generator (creates PTY) |
| `make run-edge` | Edge daemon (`SERIAL_PORT` required) |
| `make run-cloud` | Cloud ingest + HTTP API `:8080` |
| `make run-dashboard` | Vite dashboard `:5173` |
| `make test` | Go unit tests |

## Contracts

- MQTT topic: `telemetry/{well_id}/points` (QoS 1), JSON `TelemetryPoint`
- REST: `GET /api/v1/wells/{wellID}/telemetry`, `GET /api/v1/health`
- WebSocket: `GET /api/v1/ws/telemetry?well_id=...`

## Local services

| Service | Endpoint | Credentials |
|---------|----------|-------------|
| TimescaleDB | `localhost:5433` / db `telemetry` | `telemetry` / `telemetry` |
| MQTT | `localhost:1883`, WS `:9001` | anonymous (dev) |
| Cloud API | `http://localhost:8080` | — |
| Dashboard | `http://localhost:5173` | — |
