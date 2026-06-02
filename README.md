# logarift

Logarift is a local-first Developer Experience friction logging system that records, scores, and analyzes interruptions, cognitive drag, workflow rifts, and recurring sources of engineering toil.

## MVP-2 backend persistence

The backend now uses the MongoDB Go driver v2 import path (`go.mongodb.org/mongo-driver/v2`) for MongoDB connection readiness, status reporting, CQRS command/query storage flows, and collection index bootstrap. Startup fails clearly if MongoDB cannot be pinged, indexes cannot be created, command/query handlers cannot be registered, or the default MVP model configuration cannot be ensured.

### Storage architecture

The storage layer exposes explicit command and query modules under `backend/internal/store/commands`, `backend/internal/store/queries`, and `backend/internal/store/cqrs`. Commands mutate MongoDB-backed documents, queries retrieve documents, and the local dispatcher routes registered typed messages without adding external CQRS or mediator dependencies.

## Configuration

Copy `.env.example` and adjust local values as needed. The MongoDB URI is never exposed through HTTP status responses.

- `LOGARIFT_API_HOST`
- `LOGARIFT_API_PORT`
- `LOGARIFT_MONGODB_URI`
- `LOGARIFT_MONGODB_DATABASE`
- `LOGARIFT_MONGODB_CONNECT_TIMEOUT_MS` (optional, default `5000`)
- `LOGARIFT_MATH_ENGINE_PATH`
- `LOGARIFT_EXPORT_DIR`
- `LOGARIFT_READINESS_TIMEOUT_MS`
- `LOGARIFT_SHUTDOWN_TIMEOUT_MS`

## Development

Download Go modules before building when dependencies have changed:

```bash
cd backend
go mod download
```

Run the normal unit test suite:

```bash
cd backend
go test ./...
```

Optional repository integration tests are behind the `integration` build tag and are intended for a local MongoDB instance:

```bash
cd backend
LOGARIFT_MONGODB_URI=mongodb://localhost:27017 \
LOGARIFT_MONGODB_DATABASE=logarift_test \
go test -tags=integration ./...
```

Run the local Docker stack:

```bash
docker compose up --build
```
