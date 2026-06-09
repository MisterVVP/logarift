# Logarift

Logarift is an anonymous Developer Experience friction logging system for centrally deployed tech-organization use. It records, scores, and analyzes interruptions, cognitive drag, workflow rifts, and recurring sources of engineering toil without building a surveillance or individual productivity tracking tool.

The core idea:

```text
Friction is not only a logged event.
Friction is a compounding signal that affects cognitive load, flow stability, and systemic delivery drag.
```

## Implemented Scope in This Package

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
- Docker Compose local stack for DevEx platform developers and contributors
- Helm chart for centralized Kubernetes deployment across a tech organization
- optional local LLM adapter service for quick-event enrichment behind deterministic fallback

The LLM adapter is disabled unless `LOGARIFT_LLM_ADAPTER_ENABLED=true`; see `docs/technical/04_local_llm_adapter_setup.md` for Ollama/Qwen setup and optional Logarift-specific Modelfiles.

Out of scope remains:

- per-developer accounts or authorization models
- SSO enforcement such as Entra ID, AWS IAM Identity Center, Google Cloud Identity, or generic OIDC/SAML
- cloud sync controlled by the application
- team dashboards that reveal private timelines or individual rankings
- hidden telemetry
- IDE/chat/calendar ingestion
- individual productivity ranking
- AI recommendations
- LLM/ML organisation and team inference for locating systemic friction

## Deployment Model

Logarift is now positioned around a centralized private deployment model: run it as containers or install it into Kubernetes so every developer, technical lead, Developer Experience engineer, and engineering manager in a tech organization can log and inspect friction with minimal access barriers. The application should stay anonymous by default: no Logarift-owned user concept, no per-person authorization rules, and no individual productivity views.

Local Docker Compose remains important, but primarily for DevEx platform developers, contributors, demos, and safe testing before cluster rollout.

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

Run the full Docker stack with MongoDB, Valkey Streams, math engine, optional LLM adapter, backend, and frontend:

```bash
docker compose up --build
```

Open:

```text
Backend:  http://localhost:8080/api/v1/status
Frontend: http://localhost:5173
LLM adapter: http://localhost:8091/health/live
```

## Kubernetes Deployment

A Helm chart is available under `charts/logarift` for Kubernetes installs. Published releases are also available as OCI Helm charts at `oci://ghcr.io/mistervvp/charts/logarift`, alongside Docker images under `ghcr.io/mistervvp`. By default the chart deploys the frontend, backend, math engine, MongoDB, and Valkey. MongoDB and Valkey can be disabled when using externally managed services:

```bash
helm upgrade --install logarift charts/logarift \
  --set mongodb.enabled=false \
  --set mongodb.external.uri='mongodb://mongo.example:27017' \
  --set valkey.enabled=false \
  --set valkey.external.url='redis://valkey.example:6379'
```

Most Kubernetes placement controls are optional and configurable per component, including `nodeSelector`, `affinity`, pod anti-affinity through `affinity`, `tolerations`, and `topologySpreadConstraints`. The chart also supports existing Secrets for MongoDB and Valkey connection strings, persistence settings, probes, resources, Gateway API HTTPRoutes, optional LLM adapter deployment, and optional chart-managed Ollama runtime deployment.

Install a published chart from GHCR by version:

```bash
helm upgrade --install logarift oci://ghcr.io/mistervvp/charts/logarift \
  --version 0.1.0
```

Container images are published to `ghcr.io/mistervvp/logarift-api`, `ghcr.io/mistervvp/logarift-frontend`, `ghcr.io/mistervvp/logarift-math-engine`, and `ghcr.io/mistervvp/logarift-llm-adapter`. The source chart pins the stable `0.1.0` image tag for these application images; packaged release and `dev-*` charts are rewritten by the release workflow to point at the matching release or branch image tag. GitHub Releases include the packaged Helm chart, repository source archives, and checksums; pushes to `dev-*` branches publish development image tags and development chart versions for pre-release testing. See `docs/runbooks/release_packages.md` for the complete publishing and install workflow.

### Local Kubernetes quick start

For MicroK8s, enable DNS, hostpath storage, Helm, and the MicroK8s routing addon. The addon name is `ingress`, but current MicroK8s installs include Gateway API support and expose a Traefik Gateway that this chart can attach to without creating Kubernetes Ingress objects.

Use `charts/logarift/values.local.yaml` for local MicroK8s installs. It enables the optional chart-managed Ollama runtime and LLM adapter, stores local data on `microk8s-hostpath`, and keeps the adapter's Ollama traffic on Kubernetes DNS instead of LAN IPs, `localhost`, or `host.docker.internal`. The Ollama init container pulls `qwen3:8b` and creates the default `logarift-enricher-qwen3-8b` model alias from the bundled Logarift Modelfile before the Ollama container starts.

```bash
microk8s status --wait-ready
microk8s enable dns hostpath-storage helm3 ingress
microk8s helm3 upgrade --install logarift charts/logarift \
  --create-namespace --namespace logarift \
  --values charts/logarift/values.local.yaml
microk8s kubectl -n logarift rollout status statefulset/logarift-ollama
microk8s kubectl -n logarift rollout status deploy/logarift-llm-adapter
```

Add `127.0.0.1 logarift.local` to your workstation hosts file if your MicroK8s routing setup does not already resolve that name, then open `http://logarift.local`.

For Minikube, kind, Docker Desktop Kubernetes, and similar local clusters without a Gateway API controller, install the chart with defaults and port-forward the frontend Service:

```bash
kubectl create namespace logarift
helm upgrade --install logarift charts/logarift --namespace logarift
kubectl -n logarift rollout status deploy/logarift-frontend
kubectl -n logarift port-forward svc/logarift-frontend 5173:5173
```

Open `http://localhost:5173`. If your local cluster has a Gateway API implementation, use the same `gateway.enabled=true` chart values and point `httpRoute.parentRefs` at that implementation's Gateway.

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
GET    /api/v1/enrichment-jobs/{id}
GET    /api/v1/enrichment-jobs/{id}/events  # Server-Sent Events stream
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
docs/technical/02_deterministic_enrichment_engine.md
docs/technical/03_local_llm_adapter.md
docs/technical/04_local_llm_adapter_setup.md
docs/technical/05_local_ml_classifier_service.md
docs/technical/system-design.md
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

Docker Compose starts local MongoDB and Valkey services for the backend. MongoDB is the auditable data/job state store; Valkey Streams deliver asynchronous LLM enrichment jobs between the API request path and the backend worker.
