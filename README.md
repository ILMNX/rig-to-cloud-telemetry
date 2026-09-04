# Rig-to-Cloud Telemetry Monorepo

Polyglot monorepo for a resilient rig-to-cloud drilling telemetry pipeline with WITS/serial emulation.

## Layout

| Path | Role |
|------|------|
| `edge-daemon/` | Go edge agent — serial/WITS ingest, SQLite buffer, MQTT sync |
| `cloud-backend/` | Go cloud ingest — MQTT → TimescaleDB, HTTP/WebSocket API |
| `shared/` | Shared Go telemetry models (`rigtelemetry/shared`) |
| `simulator/` | Python WITS frame generator + network impairment helpers |
| `dashboard/` | React + Vite + Tailwind + ECharts well-log visualization |
| `infrastructure/` | Mosquitto & TimescaleDB init configs |

## Prerequisites

- Go 1.22+
- Python 3.10+
- Node.js 20+ / npm
- Docker & Docker Compose

## Quick start

```bash
make init        # Go mods, Python venv, npm install
make infra-up    # TimescaleDB (:5432) + Mosquitto (:1883, :9001)
make run-sim     # WITS generator
make run-edge    # Edge daemon stub
make run-cloud   # Cloud ingest stub
```

Stop infrastructure:

```bash
make infra-down
```

## Local services

| Service | Endpoint | Credentials |
|---------|----------|-------------|
| TimescaleDB | `localhost:5432` / db `telemetry` | `telemetry` / `telemetry` |
| MQTT | `localhost:1883` (MQTT), `localhost:9001` (WebSockets) | anonymous (dev only) |

## Go workspace

`go.work` links `shared`, `edge-daemon`, and `cloud-backend`. Modules:

- `rigtelemetry/shared`
- `rigtelemetry/edge` (replace → `../shared`)
- `rigtelemetry/cloud` (replace → `../shared`)
