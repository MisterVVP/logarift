# Local-First Architecture

## Purpose

This document defines the MVP local-first architecture.

The system should be easy to run locally and should not require cloud services.

## Components

MVP components:

```text
React + Vite frontend
Go backend API
MongoDB
C++ math engine service
Docker Compose
```

## Runtime Topology

```text
Browser
  |
  v
React + Vite frontend
  |
  v
Go backend API
  |
  +--> MongoDB
  |
  +--> C++ math engine service
```

## Primary Event Flow

```text
User logs event in frontend
Frontend sends request to Go API
Go API validates event
Go API stores event in MongoDB
Go API requests score calculation
C++ math engine service receives JSON input over HTTP
C++ math engine service returns JSON score output
Go API stores score snapshot
Frontend displays dashboard
```

## Backend Responsibilities

The Go backend owns:

- HTTP API
- request validation
- MongoDB access
- event persistence
- session persistence
- goal persistence
- score snapshot persistence
- export generation
- calling the C++ math engine service
- serving frontend in packaged mode if needed later

## Frontend Responsibilities

The React + Vite frontend owns:

- friction event form
- session/goal UI
- event timeline
- filters
- dashboard charts
- export controls
- local settings UI

## MongoDB Responsibilities

MongoDB stores:

- friction events
- work sessions
- work goals
- score snapshots
- model configs
- export metadata

MongoDB should run locally through Docker Compose in MVP.

## C++ Math Engine Responsibilities

The C++ math engine owns deterministic score calculation.

The engine should:

- read JSON input
- validate required fields
- compute MVP scores
- return JSON output
- avoid hidden state
- be testable independently

## Why a Separate Math Service

The C++ engine runs as a separate local service instead of being linked into the Go backend.

Reasons:

- avoids cgo complexity
- keeps the language boundary explicit
- allows independent scaling and testing
- keeps backend and scoring failures isolated
- preserves deterministic JSON input/output
- fits Docker Compose local development

The same binary may retain CLI-compatible stdin/stdout mode for smoke tests, but backend integration should use HTTP.

## API Boundary

The Go backend communicates with the C++ engine over HTTP.

Preferred MVP approach:

```text
Go serializes scoring request to JSON
Go posts JSON to math-engine /v1/score
Math engine returns JSON response
Go parses JSON response
Go stores score snapshot
```

## Local Development

Expected local development command:

```bash
docker compose up --build
```

Expected services:

```text
backend
frontend
mongodb
math-engine
```

The C++ math engine has its own Docker image and runs as a separate local service.

## Configuration

Configuration should use environment variables.

Suggested variables:

```text
LOGARIFT_API_PORT
LOGARIFT_MONGODB_URI
LOGARIFT_MONGODB_DATABASE
LOGARIFT_MATH_ENGINE_URL
LOGARIFT_EXPORT_DIR
```

## Security Assumptions

MVP is local-only.

Do not expose MongoDB publicly.

Do not bind services to public interfaces unless explicitly configured.

Do not collect hidden telemetry.

## Failure Handling

If MongoDB is unavailable:

- backend health check should fail
- API should return clear error

If math engine fails:

- raw friction events should still be saved
- score calculation should return clear error
- failed score attempts should not corrupt existing score snapshots

If frontend cannot reach backend:

- show a clear local connection error

## Future Architecture Extensions

Possible later additions:

- Git importer
- CI importer
- plugin runner
- local team aggregation server
- advanced math service
- report generator
- backup/restore service

## Quick Logging Enrichment Flow

The backend includes an in-process deterministic enrichment engine for quick logging.

```text
User enters date, friction level, and notes
Frontend sends POST /api/v1/friction-events/quick
Go API validates the observed fields
Go deterministic enrichment engine infers canonical event fields
Go API stores observed/inferred/canonical event data in MongoDB
Dashboard and math-engine scoring use canonical fields
```

The enrichment engine is intentionally separate from the C++ math engine:

```text
enrichment engine = interpretation of notes into structured fields
math engine       = deterministic scoring from canonical structured fields
```

Future local LLM or local ML adapters may be added behind the enrichment boundary without changing the math-engine contract.

## Upload flow

For rich notes, the frontend may upload screenshots to the backend through `POST /api/v1/uploads`. The backend writes accepted images to `LOGARIFT_UPLOAD_DIR` and returns a local `/uploads/{filename}` URL. The notes editor inserts that URL into the event notes. No cloud object storage is used in MVP.

## Math Engine Observability

The C++ math engine emits structured JSON logs in service mode. Logs include score request lifecycle events and score calculation summaries with event count, period, CLA, FCI, SDC, wait minutes, active minutes, and duration.
