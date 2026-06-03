# CQRS Backend Architecture

## Summary

The backend uses a small in-process CQRS dispatcher as the application boundary between product services and persistence. HTTP handlers call domain services, services validate and normalize request DTOs, and services then send command/query messages through the shared dispatcher.

Services must not depend on MongoDB repositories directly.

## Request flow

```text
HTTP handler
  -> service validation/defaults/normalization
  -> cqrs.Dispatcher.SendCommand or SendQuery
  -> repository-backed command/query handler registered by mongostore.Store
  -> MongoDB collection operation
```

## Singleton dispatcher

The API process creates one dispatcher during startup:

```go
dispatcher := cqrs.NewDispatcher()
stores := mongostore.New(db)
stores.RegisterCQRS(dispatcher)
```

`RegisterCQRS` validates that all required private repositories are present before any handler is registered. Individual handler methods do not perform repeated nil repository checks.

That dispatcher is then passed to every service through the HTTP server constructor. The dispatcher owns handler registration maps and synchronization, so creating separate dispatchers per service would bypass the intended shared command/query registry.

## Repository encapsulation

Repositories are internal persistence details of `internal/store/mongostore`. They are intentionally held in unexported `Store` fields and are not exposed through public repository interfaces or server constructors.

Only repository-backed CQRS handlers registered from `mongostore.Store.RegisterCQRS` should call repository methods. This keeps HTTP and service packages decoupled from persistence and makes direct repository usage outside the persistence/handler layer impractical.

## Message ownership

Command and query message types live in:

- `backend/internal/store/commands`
- `backend/internal/store/queries`

Those packages define message contracts only. They do not expose repositories and do not construct repository-backed handlers. Handler registration is owned by `mongostore.Store` because that package owns the concrete repositories. The dispatcher provides reflection-based registration helpers (`RegisterHandler` and `RegisterHandlers`) that inspect handler signatures and classify messages from `backend/internal/store/commands` as commands and messages from `backend/internal/store/queries` as queries. This keeps command/query registration centralized and avoids duplicating generic registration boilerplate for every message type.

## Service responsibilities

Services remain responsible for:

- request DTO validation
- ontology validation
- text normalization
- default values
- server-owned timestamps and schema version
- ObjectID string parsing for API-facing IDs
- translating persistence not-found errors into service-level not-found errors

Services then send commands for mutations and queries for reads.

## Testing guidance

Service tests should register fake command/query handlers on a real `cqrs.Dispatcher` rather than passing fake repositories to services. Handler tests may use the in-memory Mongo driver through `mongostore.Store.RegisterCQRS` to exercise the same dispatcher boundary used by production startup.
