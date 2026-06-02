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
C++ math engine CLI
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
  +--> C++ math engine CLI
```

## Primary Event Flow

```text
User logs event in frontend
Frontend sends request to Go API
Go API validates event
Go API stores event in MongoDB
Go API requests score calculation
C++ math CLI receives JSON input
C++ math CLI returns JSON score output
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
- invoking the C++ math CLI
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

## Why CLI First

The C++ engine should initially be called as a CLI executable instead of a shared library.

Reasons:

- simpler implementation
- avoids cgo complexity
- explicit language boundary
- easy process isolation
- deterministic JSON input/output
- independent testability
- easier for agentic implementation

Future versions may expose a shared library if performance requires it.

## API Boundary

The Go backend should communicate with the C++ engine using files or stdin/stdout JSON.

Preferred MVP approach:

```text
Go serializes scoring request to JSON
Go sends JSON to math CLI through stdin
Math CLI writes JSON response to stdout
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
```

The C++ math engine may be built as part of backend image or mounted as a local binary during early development.

## Configuration

Configuration should use environment variables.

Suggested variables:

```text
LOGARIFT_API_PORT
LOGARIFT_MONGODB_URI
LOGARIFT_MONGODB_DATABASE
LOGARIFT_MATH_ENGINE_PATH
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
