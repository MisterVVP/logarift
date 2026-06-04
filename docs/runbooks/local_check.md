# Local Check Runbook

## 1. Download backend dependencies

The backend uses the official MongoDB Go driver. A fresh checkout needs network access once to download Go modules.

```bash
make deps-backend
```

## 2. Run unit and math-engine tests

```bash
make test
```

Expected result:

- C++ math engine CLI-compatible sample test prints a JSON line containing `"cla"`.
- Go backend tests pass.

## 3. Build local binaries

```bash
make build
```

Expected binaries:

```text
bin/logarift-api
bin/logarift-math-engine
```

## 4. Run math engine and backend directly

Direct backend execution expects MongoDB at `mongodb://localhost:27017`. Use Docker Compose for the easiest full local stack, or start MongoDB separately before running the backend directly.

Terminal 1:

```bash
make run-math
```

Terminal 2:

```bash
make run-backend
```

Then in another terminal:

```bash
curl http://localhost:8080/health/live
curl http://localhost:8080/health/ready
curl http://localhost:8080/api/v1/status
```

## 5. Create sample data

Create a goal:

```bash
curl -X POST http://localhost:8080/api/v1/work-goals \
  -H "Content-Type: application/json" \
  -d '{"title":"Implement dashboard","status":"active"}'
```

Create a session:

```bash
curl -X POST http://localhost:8080/api/v1/work-sessions \
  -H "Content-Type: application/json" \
  -d '{"title":"Morning work","started_at":"2026-06-01T08:30:00Z"}'
```

Create a friction event:

```bash
curl -X POST http://localhost:8080/api/v1/friction-events \
  -H "Content-Type: application/json" \
  -d '{
    "timestamp_start":"2026-06-01T09:15:00Z",
    "workflow_stage":"test",
    "friction_layer":"technical",
    "friction_type":"failed_feedback",
    "severity_self":4,
    "cognitive_load_self":3,
    "emotion_valence":-1,
    "time_lost_minutes":20,
    "resume_time_minutes":8,
    "interruption_count":1,
    "tags":["ci","flaky-test"],
    "notes":"CI failed with a test that passed locally."
  }'
```

Calculate scores:

```bash
curl -X POST http://localhost:8080/api/v1/scores/calculate \
  -H "Content-Type: application/json" \
  -d '{
    "period_start":"2026-06-01T00:00:00Z",
    "period_end":"2026-06-07T23:59:59Z",
    "score_type":"weekly"
  }'
```

List score snapshots:

```bash
curl http://localhost:8080/api/v1/score-snapshots
```

## 6. Run frontend directly

```bash
cd frontend
npm install
npm run dev
```

Open:

```text
http://localhost:5173
```

## 7. Run Docker Compose

Docker Compose starts four services: MongoDB, the C++ math engine service, the Go backend, and the React/Vite frontend.

```bash
docker compose up --build
```

Open:

```text
Backend:  http://localhost:8080/api/v1/status
Frontend: http://localhost:5173
```

Stop:

```bash
docker compose down
```

## Docker module download behavior

The backend Dockerfile copies the backend source first and then runs `go mod download`. This is intentional: it allows Docker to generate `go.sum` inside the build stage even if the local archive has an empty `backend/go.sum`.
