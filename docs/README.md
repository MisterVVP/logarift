# Logarift Documentation

## Overview

Logarift is a local-first Developer Experience friction logging and analysis system.

The project is based on a simple idea:

```text
Friction is not only a logged event.
Friction is a compounding signal that affects cognitive load, flow stability, and systemic delivery drag.
```

The documentation in this folder is split into:

- `product/` for product foundations, initial release boundaries, ontology, data model, local-first principles, and privacy/IP governance.
- `technical/` for implementation architecture decisions and backend/frontend/persistence/math-engine technical notes.

initial planning baseline is intentionally documentation-only. It should guide later implementation tasks for the Go backend, React + Vite frontend, MongoDB persistence layer, and C++ math engine.

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

#### 4. Initial Release Scope

[product/03_initial_scope.md](./product/03_initial_scope.md)

Defines what is included and excluded from the initial release.

The initial release includes local single-user mode, manual friction logging, goals, sessions, MongoDB persistence, basic dashboarding, C++ scoring service integration, JSON export, and seed/demo data.

The initial release excludes cloud deployment, authentication, team dashboards, IDE plugins, external importers, hidden telemetry, advanced research models, and AI-generated recommendations.

#### 5. Math Model

[product/04_math_model.md](./product/04_math_model.md)

Defines the first deterministic scoring model.

initial release scores include:

- Cognitive Load Accumulator
- Friction Compounding Index
- Systemic Drag Coefficient
- Friction Cost Score

The formulas are product hypotheses, not validated universal scientific metrics. They should be implemented in a transparent and explainable way.

#### 6. MongoDB Data Model

[product/05_mongodb_data_model.md](./product/05_mongodb_data_model.md)

Defines the initial release MongoDB collections and document shapes.

initial release collections:

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

Defines the high-level initial release architecture.

initial release components:

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

#### 9. Quick Logging and Enrichment

[product/08_quick_logging_and_enrichment.md](./product/08_quick_logging_and_enrichment.md)

Defines the new three-field logging UX, color-coded friction levels, observed/inferred/canonical event model, deterministic rule engine, and future local LLM/ML extension points.

### Technical documents

#### 1. CQRS Backend Architecture

[technical/01_cqrs_backend_architecture.md](./technical/01_cqrs_backend_architecture.md)

Defines the backend CQRS boundary used by the Go API. Services send command/query messages through one shared dispatcher; repository access is private to the Mongo store and its registered CQRS handlers.

#### 2. Deterministic Enrichment Engine

[technical/02_deterministic_enrichment_engine.md](./technical/02_deterministic_enrichment_engine.md)

Documents the local rules engine used by `POST /api/v1/friction-events/quick` and explains how it remains separate from the C++ math engine.

#### 3. Local LLM Adapter

[technical/03_local_llm_adapter.md](./technical/03_local_llm_adapter.md)

Defines the optional local-only LLM adapter service boundary, Ollama-compatible runtime contract, backend merge policy, privacy constraints, and setup requirements.

#### 4. Local LLM Adapter Setup

[technical/04_local_llm_adapter_setup.md](./technical/04_local_llm_adapter_setup.md)

Documents Ubuntu and Windows 11 setup for Ollama and Qwen, adapter environment variables, smoke requests, and official upstream references.

#### 5. Local ML Classifier Service

[technical/05_local_ml_classifier_service.md](./technical/05_local_ml_classifier_service.md)

Defines the future optional local-only classifier service boundary, training data source, inference contract, correction workflow, and relationship to deterministic and LLM enrichment.

#### 6. System Design

[technical/system-design.md](./technical/system-design.md)

Living system-design document for runtime component boundaries, asynchronous LLM enrichment, backend worker integration, UI SSE updates, scoring, observability, and Docker Compose/Kubernetes deployment.

#### 7. Local Check Runbook

[runbooks/local_check.md](./runbooks/local_check.md)

Step-by-step commands for testing the backend, math engine, scoring endpoint, frontend, and Docker Compose stack.


## Design Constraints

Implementation work should preserve these constraints:

- MongoDB is the initial release persistence backend.
- The system is local-first.
- The initial release is single-user.
- The initial release does not include authentication.
- The initial release does not include cloud sync.
- The initial release does not include hidden telemetry.
- The initial release does not rank developers.
- The initial release math is deterministic and explainable.
- The default logging UI uses three fields and relies on local deterministic enrichment for structured fields.
- The initial release C++ math engine is a separate service in Docker Compose and remains CLI-compatible for local deterministic tests.
- Proprietary company taxonomies, survey wording, dashboards, and scoring systems must not be copied.

## Contribution Notes

When adding or changing documentation:

- keep initial release and future scope clearly separated
- prefer precise implementation-oriented language
- preserve original project terminology unless intentionally evolving it
- update this README when adding new documentation files
- document major product or architecture changes before implementing them
