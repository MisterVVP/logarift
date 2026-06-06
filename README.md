# Logarift

Logarift is a local-first Developer Experience friction logging system that records, scores, and analyzes interruptions, cognitive drag, workflow rifts, and recurring sources of engineering toil.

The core idea:

```text
Friction is not only a logged event.
Friction is a compounding signal that affects cognitive load, flow stability, and systemic delivery drag.
```

## Implemented MVP Scope in This Package

This package includes MVP-3 through MVP-7 on top of the existing MVP-0/MVP-1/MVP-2 foundation:

- three-field quick friction logging with deterministic local enrichment
- manual CRUD APIs for friction events, work goals, and work sessions
- deterministic C++ math engine service with CLI-compatible mode
- Go backend scoring integration over HTTP (`LOGARIFT_MATH_ENGINE_URL`)
- persisted score snapshots
- React + Vite local logging UI
- simplified two-tab UI: quick logging/recent logs first, dashboard second
- rich notes editor with formatted text, links, pasted screenshots, and local image uploads
- dashboard cards and breakdowns with tooltips
- structured math-engine calculation logs
- Docker Compose local stack
- optional local LLM adapter service for quick-event enrichment behind deterministic fallback

The LLM adapter is disabled unless `LOGARIFT_LLM_ADAPTER_ENABLED=true`; see `docs/technical/04_local_llm_adapter_setup.md` for Ollama/Qwen setup.

Out of scope remains:

- authentication
- cloud sync
- team dashboards
- hidden telemetry
- IDE/chat/calendar ingestion
- individual productivity ranking
- AI recommendations

## Repository Layout

```text
backend/       Go backend API
frontend/      React + Vite frontend
math-engine/   C++ scoring service and CLI-compatible scorer
llm-adapter/   Optional Go service that calls local Ollama-compatible models
docs/          Product, technical, and runbook docs
exports/       Local export target placeholder
data/uploads/  Local uploaded screenshots when running outside Docker
scripts/       Convenience scripts
```

## Requirements

For direct local development:

- Go 1.25
- C++17 compiler such as `g++`
- Node.js 20+ or 22+
- npm

For Docker:

- Docker
- Docker Compose

Docker backend builds use `golang:1.25`.

### Go module dependency note

The backend uses the official MongoDB Go driver v2. Docker builds run `go mod download` inside the `golang:1.25` build stage after copying the backend source, so a missing or empty local `backend/go.sum` does not break `docker compose up --build`. For local non-Docker builds, run `make deps-backend` once with Go 1.25 available.

## Quick Check

Run backend and math tests:

```bash
make test
```

Build local binaries:

```bash
make build
```

Run the math engine and backend directly in separate terminals. Direct backend execution also expects MongoDB at `mongodb://localhost:27017`:

```bash
make run-math
```

```bash
make run-backend
```

Check status:

```bash
curl http://localhost:8080/api/v1/status
```

Run frontend directly:

```bash
cd frontend
npm install
npm run dev
```

Open:

```text
http://localhost:5173
```

Run the full Docker stack with MongoDB, math engine, optional LLM adapter, backend, and frontend:

```bash
docker compose up --build
```

Open:

```text
Backend:  http://localhost:8080/api/v1/status
Frontend: http://localhost:5173
LLM adapter: http://localhost:8091/health/live
```

## API Overview

Health/status:

```text
GET /health/live
GET /health/ready
GET /api/v1/status
```

Uploads:

```text
POST /api/v1/uploads
GET  /uploads/{filename}
```

Friction events:

```text
POST   /api/v1/friction-events/quick
POST   /api/v1/friction-events
GET    /api/v1/friction-events
GET    /api/v1/friction-events/{id}
PUT    /api/v1/friction-events/{id}
DELETE /api/v1/friction-events/{id}
```

Quick event example. The UI uses the same endpoint after uploading any pasted or attached screenshots to `/api/v1/uploads`:

```bash
curl -X POST http://localhost:8080/api/v1/friction-events/quick \
  -H "Content-Type: application/json" \
  -d '{"occurred_at":"2026-06-04T19:26:00Z","friction_level":"orange","notes_markdown":"CI failed again after 20 min with an unclear timeout."}'
```

Work goals:

```text
POST   /api/v1/work-goals
GET    /api/v1/work-goals
GET    /api/v1/work-goals/{id}
PUT    /api/v1/work-goals/{id}
DELETE /api/v1/work-goals/{id}
```

Work sessions:

```text
POST   /api/v1/work-sessions
GET    /api/v1/work-sessions
GET    /api/v1/work-sessions/{id}
PUT    /api/v1/work-sessions/{id}
DELETE /api/v1/work-sessions/{id}
```

Scoring:

```text
POST /api/v1/scores/calculate
GET  /api/v1/score-snapshots
GET  /api/v1/score-snapshots/{id}
```

## Documentation

Start with:

```text
docs/README.md
```

Useful implementation docs:

```text
docs/product/08_quick_logging_and_enrichment.md
docs/technical/02_mvp_3_to_7_implementation.md
docs/technical/03_deterministic_enrichment_engine.md
docs/runbooks/local_check.md
math-engine/README.md
frontend/README.md
```

## Storage and Math Engine Notes

The backend uses the official MongoDB Go driver v2 package directly. No local MongoDB driver shim is included.

The math engine runs as a separate C++ HTTP service in Docker Compose. The backend calls it through `LOGARIFT_MATH_ENGINE_URL`, which is `http://math-engine:8090` in Docker Compose and `http://localhost:8090` for direct local runs. The math engine writes structured JSON logs for server startup, score requests, calculation summaries, status, duration, event count, CLA, FCI, SDC, wait minutes, and active minutes.

When setting up a fresh checkout, run:

```bash
cd backend
go mod download
```

Docker Compose starts a local MongoDB service for the backend.
