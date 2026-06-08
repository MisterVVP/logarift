# Initial Release Scope

## Purpose

This document defines the initial release boundary.

The initial release should validate the core product concept with the smallest useful local-first implementation.

## Product Statement

The initial release allows a single local user to manually log Developer Experience friction, classify it with the original ontology, compute basic mathematical scores, and inspect local analytics through a dashboard.

## Included in Initial Release

### Local Single-User Mode

The initial release runs locally and assumes one user.

No authentication is required.

### Manual Friction Event Logging

The user can create, edit, list, filter, and delete friction events.

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

The user can define work goals and link friction events to them.

Example goals:

- implement feature
- fix bug
- review pull request
- debug issue

### Session Tracking

The user can create work sessions and link events to sessions.

### MongoDB Persistence

The initial release uses MongoDB as the local data store.

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

The dashboard should show:

- events over time
- friction by workflow stage
- friction by layer
- friction by type
- top time-loss sources
- top cognitive-load sources
- score cards for initial release metrics

### C++ Scoring Service

The initial release includes a deterministic C++ scoring application. In Docker Compose it runs as a separate HTTP service. For local tests it also supports CLI-compatible stdin/stdout mode.

The service receives JSON input and returns JSON output.

### Go API Integration with Scoring Service

The Go backend calls the C++ scoring service and stores score snapshots.

### JSON Export

The user can export events and score snapshots as JSON.

### Seed/Demo Dataset

The repository includes a small sample dataset for local testing and demo mode.

## Excluded from Initial Release

The following are explicitly out of scope:

```text
cloud deployment
multi-user authentication
team dashboards
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
plugin system
mobile application
```

## Non-Goals

The initial release must not attempt to prove causal relationships.

The initial release must not claim that the mathematical scores are scientifically validated universal metrics.

The initial release must not rank developers.

The initial release must not silently collect telemetry.

The initial release must not require cloud services.

## Success Criteria

The initial release is successful when:

- local setup works with Docker Compose
- the user can log a friction event in under 15 seconds
- events are persisted in MongoDB
- the dashboard shows basic analytics
- the C++ scoring service produces deterministic output
- the Go backend can call the scoring service
- the user can export data as JSON
- the system remains understandable and explainable

## Future Stages

After initial release, possible stages include:

1. local Git and CI importers
2. advanced C++ math engine
3. insight and recommendation engine
4. intervention simulation
5. optional local team mode
6. advanced research models
7. extensibility and plugins
