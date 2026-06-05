# Logarift Math Engine

The MVP math engine is a small deterministic C++ application. It can run in two modes:

- HTTP service mode for Docker Compose and backend integration.
- CLI-compatible mode for local smoke tests, where it reads JSON from stdin and writes JSON to stdout.

## Build

```bash
make -C math-engine
```

The default binary is:

```text
bin/logarift-math-engine
```

## Run as service

```bash
LOGARIFT_MATH_ENGINE_PORT=8090 ./bin/logarift-math-engine --serve
```

Health endpoints:

```text
GET /health/live
GET /health/ready
```

Scoring endpoint:

```text
POST /v1/score
```

## Smoke test CLI-compatible mode

```bash
make -C math-engine test
```

Direct CLI-compatible call:

```bash
./bin/logarift-math-engine < math-engine/samples/scoring-request.sample.json
```

## Docker

The math engine has its own Docker image and service in `docker-compose.yml`:

```text
math-engine
```

The backend calls it with:

```text
LOGARIFT_MATH_ENGINE_URL=http://math-engine:8090
```

## Input shape

```json
{
  "model_version": "mvp-0.1",
  "period_start": "2026-06-01T00:00:00Z",
  "period_end": "2026-06-07T23:59:59Z",
  "events": []
}
```

## Implemented MVP scores

- Cognitive Load Accumulator (`cla`)
- Friction Compounding Index (`fci`)
- Systemic Drag Coefficient (`sdc`)
- per-event Friction Cost Score (`fcs`)

The formulas are deterministic MVP product hypotheses and are not presented as universal validated scientific metrics.

## Observability

Service mode writes structured JSON logs to stderr/stdout as container logs. Logs include:

- `math engine listening` with port
- `score request received` with request id and payload size
- `score calculation completed` with event count, period, CLA, FCI, SDC, total wait minutes, total active minutes, top contributor count, and calculation duration
- `score request completed` with status and duration

Example:

```json
{"level":"info","service":"logarift-math-engine","message":"score calculation completed","event_count":"3","cla":"31.2000","fci":"12.8000","sdc":"0.4200"}
```

CLI-compatible mode still writes score JSON to stdout, so scripts can parse it. Calculation logs are emitted separately and do not change the response contract.
