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

Create a quick friction event using the default three-field UX contract:

```bash
curl -X POST http://localhost:8080/api/v1/friction-events/quick \
  -H "Content-Type: application/json" \
  -d '{
    "occurred_at":"2026-06-01T09:15:00Z",
    "friction_level":"orange",
    "notes_markdown":"CI failed again after 20 min with an unclear timeout. https://github.com/org/repo/actions/runs/123"
  }'
```

The backend stores the three observed fields plus inferred and canonical fields for dashboarding and scoring. The full advanced endpoint remains available at `POST /api/v1/friction-events`.

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

## 8. Check rich notes upload

The frontend supports the easiest path: paste a screenshot into the notes editor or click **Screenshot** and select an image.

Backend upload endpoint smoke test:

```bash
base64 -d > /tmp/logarift-test.png <<'PNG'
iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO+/p9sAAAAASUVORK5CYII=
PNG
curl -X POST http://localhost:8080/api/v1/uploads \
  -F "file=@/tmp/logarift-test.png;type=image/png"
```

The response contains a local `url_path` such as:

```json
{
  "url_path": "/uploads/example.png"
}
```

Uploaded images are stored under `LOGARIFT_UPLOAD_DIR` and are served from `/uploads/{filename}`.

## 9. Check math-engine logs

The math engine now emits structured JSON logs. In Docker Compose, run:

```bash
docker compose logs -f math-engine
```

Then calculate scores from the dashboard or API. Expected log messages include:

```text
math engine listening
score request received
score calculation completed
score request completed
```

The calculation log includes event count, period, CLA, FCI, SDC, wait minutes, active minutes, and duration.
