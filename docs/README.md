# Logarift Documentation

## Overview

Logarift is a local-first Developer Experience friction logging and analysis system.

The project is based on a simple idea:

```text
Friction is not only a logged event.
Friction is a compounding signal that affects cognitive load, flow stability, and systemic delivery drag.
```

The documentation in this folder is split into:

- `product/` for product foundations, MVP boundaries, ontology, data model, local-first principles, and privacy/IP governance.
- `technical/` for implementation architecture decisions and backend/frontend/persistence/math-engine technical notes.

MVP-0 is intentionally documentation-only. It should guide later implementation tasks for the Go backend, React + Vite frontend, MongoDB persistence layer, and C++ math engine.

## Documentation Map

Read the product documents in this order, then consult technical documents when implementing or reviewing code.

### Product documents

#### 1. Product Vision

[product/00_product_vision.md](./product/00_product_vision.md)

Defines the product direction, target users, core idea, differentiation, local-first principle, and long-term vision.

Start here to understand what Logarift is and what it should avoid becoming.

#### 2. Research Foundation

[product/01_research_foundation.md](./product/01_research_foundation.md)

Summarizes the public research and scientific concepts behind the product direction.

Covers cognitive load, interruptions, flow, motivation, queueing theory, Bayesian updating, Markov/state-transition modeling, DORA, and SPACE.

This document is background only. The project still uses its own original terminology, ontology, schemas, formulas, and implementation.

#### 3. Original Friction Ontology

[product/02_original_friction_ontology.md](./product/02_original_friction_ontology.md)

Defines the project’s initial original vocabulary for friction analysis.

Includes:

- goals
- sessions
- friction events
- friction episodes
- workflow stages
- friction layers
- friction types
- outcomes
- recovery signals

Use this document when implementing event models, validation, UI dropdowns, and analytics grouping.

#### 4. MVP Scope

[product/03_mvp_scope.md](./product/03_mvp_scope.md)

Defines what is included and excluded from the MVP.

The MVP includes local single-user mode, manual friction logging, goals, sessions, MongoDB persistence, basic dashboarding, C++ scoring service integration, JSON export, and seed/demo data.

The MVP excludes cloud deployment, authentication, team dashboards, IDE plugins, external importers, hidden telemetry, advanced research models, and AI-generated recommendations.

#### 5. MVP Math Model

[product/04_mvp_math_model.md](./product/04_mvp_math_model.md)

Defines the first deterministic scoring model.

MVP scores include:

- Cognitive Load Accumulator
- Friction Compounding Index
- Systemic Drag Coefficient
- Friction Cost Score

The formulas are product hypotheses, not validated universal scientific metrics. They should be implemented in a transparent and explainable way.

#### 6. MongoDB Data Model

[product/05_mongodb_data_model.md](./product/05_mongodb_data_model.md)

Defines the MVP MongoDB collections and document shapes.

MVP collections:

```text
friction_events
work_sessions
work_goals
score_snapshots
model_configs
exports
```

Use this document when implementing repositories, indexes, validation, seed data, and API payload mapping.

#### 7. Local-First Architecture

[product/06_local_first_architecture.md](./product/06_local_first_architecture.md)

Defines the high-level MVP architecture.

MVP components:

```text
React + Vite frontend
Go backend API
MongoDB
C++ math engine service
Docker Compose
```

This document explains the local architecture and the explicit HTTP boundary between the Go backend and the separate C++ math engine service.

#### 8. Privacy and IP Governance

[product/07_privacy_ip_governance.md](./product/07_privacy_ip_governance.md)

Defines privacy and intellectual property principles.

Important principles:

- local-first by default
- no hidden telemetry
- no individual productivity ranking
- no employee surveillance
- private notes stay local
- original ontology, schemas, formulas, UI, docs, and implementation belong to the project
- public scientific research may inform the system, but proprietary company work must not be copied

### Technical documents

#### 1. CQRS Backend Architecture

[technical/01_cqrs_backend_architecture.md](./technical/01_cqrs_backend_architecture.md)

Defines the backend CQRS boundary used by the Go API. Services send command/query messages through one shared dispatcher; repository access is private to the Mongo store and its registered CQRS handlers.

## Design Constraints

Implementation work should preserve these constraints:

- MongoDB is the MVP persistence backend.
- The system is local-first.
- The MVP is single-user.
- The MVP does not include authentication.
- The MVP does not include cloud sync.
- The MVP does not include hidden telemetry.
- The MVP does not rank developers.
- The MVP math is deterministic and explainable.
- The MVP C++ math engine is a separate service in Docker Compose and remains CLI-compatible for local deterministic tests.
- Proprietary company taxonomies, survey wording, dashboards, and scoring systems must not be copied.

## Contribution Notes

When adding or changing documentation:

- keep MVP and future scope clearly separated
- prefer precise implementation-oriented language
- preserve original project terminology unless intentionally evolving it
- update this README when adding new documentation files
- document major product or architecture changes before implementing them


### 9. MVP-3 to MVP-7 Implementation Notes

[technical/02_mvp_3_to_7_implementation.md](./technical/02_mvp_3_to_7_implementation.md)

Describes the implemented backend CRUD APIs, C++ math engine, backend scoring integration, React + Vite frontend, and dashboard.

### 10. Local Check Runbook

[runbooks/local_check.md](./runbooks/local_check.md)

Step-by-step commands for testing the backend, math engine, scoring endpoint, frontend, and Docker Compose stack.
