# Initial Release Scope

## Purpose

This document defines the initial public release boundary.

The initial public release should validate the core product concept with the smallest useful centrally deployable implementation while preserving local Docker Compose for DevEx platform developers, contributors, demos, and safe offline testing.

## Product Statement

The initial public release allows anonymous members of a tech organization to manually log Developer Experience friction, classify it with the original ontology, compute basic mathematical scores, and inspect shared analytics through a dashboard. Logarift does not need a first-party concept of a user, ownership, or per-person authorization for this release.

## Included in Initial Release

### Centralized Anonymous Mode

The primary deployment model is a centrally operated private service for a tech organization.

The release should run as containers and through the Helm chart so platform teams can provide one low-barrier Logarift instance for developers, technical leads, Developer Experience engineers, and engineering managers.

No Logarift-managed user identity, authorization model, or individual account workflow is required. Friction events should stay anonymous at the application layer.

### Local Developer Mode

Docker Compose remains supported for DevEx platform developers, contributors, local demos, and validation before cluster rollout.

Local mode uses the same component boundaries as centralized mode so platform developers can reproduce production behavior without needing a shared cluster.

### Manual Friction Event Logging

A person can create, edit, list, filter, and delete friction events.

Each event includes:

```text
timestamp_start
timestamp_end
workflow_stage
friction_layer
friction_type
severity_self
cognitive_load_self
emotion_valence
time_lost_minutes
resume_time_minutes
interruption_count
goal_id
session_id
tags
notes
source
```

### Goal Tracking

People can define work goals and link friction events to them.

Example goals:

- implement feature
- fix bug
- review pull request
- debug issue

### Session Tracking

People can create work sessions and link events to sessions.

### MongoDB Persistence

The initial release uses MongoDB as the data store. It can run in-cluster for small installations or connect to externally managed MongoDB for centralized deployments.

Required collections:

```text
friction_events
work_sessions
work_goals
score_snapshots
model_configs
exports
```

### Basic Dashboard

The dashboard should show anonymous aggregate views:

- events over time
- friction by workflow stage
- friction by layer
- friction by type
- top time-loss sources
- top cognitive-load sources
- score cards for initial release metrics

### C++ Scoring Service

The initial release includes a deterministic C++ scoring application. In Docker Compose and Kubernetes it runs as a separate HTTP service. For local tests it also supports CLI-compatible stdin/stdout mode.

The service receives JSON input and returns JSON output.

### Go API Integration with Scoring Service

The Go backend calls the C++ scoring service and stores score snapshots.

### JSON Export

A deployment operator or authorized infrastructure path can export events and score snapshots as JSON. The product UI must not turn export into a way to inspect individual private timelines.

### Seed/Demo Dataset

The repository includes a small sample dataset for local testing and demo mode.

## Excluded from Initial Release

The following are explicitly out of scope:

```text
Logarift-managed user accounts
per-person authorization rules
SSO enforcement
team dashboards that expose individual timelines
individual productivity ranking
IDE plugins
GitHub/GitLab apps
Jira/Linear importers
chat/calendar ingestion
background telemetry
advanced Bayesian inference
full Markov model
hidden Markov model
advanced survival analysis
AI-generated recommendations
LLM/ML organisation and team inference
plugin system
mobile application
```

## Non-Goals

The initial release must not attempt to prove causal relationships.

The initial release must not claim that the mathematical scores are scientifically validated universal metrics.

The initial release must not rank developers.

The initial release must not identify who experienced or caused a friction event.

The initial release must not silently collect telemetry.

The initial release must not require a Logarift cloud service.

## Success Criteria

The initial release is successful when:

- centralized setup works with containers or the Helm chart
- local setup works with Docker Compose
- a person can log a friction event in under 15 seconds
- events are persisted in MongoDB without Logarift-managed user identity
- the dashboard shows basic anonymous analytics
- the C++ scoring service produces deterministic output
- the Go backend can call the scoring service
- export paths are explicit and privacy-aware
- the system remains understandable and explainable

## Future Stages

After initial release, possible stages include:

1. optional access gate through SSO such as Microsoft Entra ID, AWS IAM Identity Center, Google Cloud Identity, or generic OIDC/SAML
2. LLM/ML-assisted friction location that suggests likely systems, teams, or organization areas from anonymous friction logs
3. local Git and CI importers
4. advanced C++ math engine
5. insight and recommendation engine
6. intervention simulation
7. anonymous aggregate team and organisation views with minimum cohort protections
8. advanced research models
9. extensibility and plugins
