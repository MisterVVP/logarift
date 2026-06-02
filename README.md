# logarift

Logarift is a local-first Developer Experience friction logging system that records, scores, and analyzes interruptions, cognitive drag, workflow rifts, and recurring sources of engineering toil.

## MVP-2 backend persistence

The backend now uses the MongoDB Go driver v2 import path (`go.mongodb.org/mongo-driver/v2`) for MongoDB connection readiness, status reporting, document repositories, and collection index bootstrap. Startup fails clearly if MongoDB cannot be pinged, indexes cannot be created, or the default MVP model configuration cannot be ensured.

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
