# MongoDB Data Model

## Purpose

This document defines the initial release MongoDB data model.

MongoDB is used because friction data is naturally document-shaped and expected to evolve.

The initial release should use versioned documents rather than SQL migrations.

## General Document Rules

Every persisted document should include:

```text
schema_version
created_at
updated_at
```

Timestamps should be stored in UTC.

IDs should use MongoDB ObjectId by default unless a stable external ID is needed.

## Collections

initial release collections:

```text
friction_events
work_sessions
work_goals
score_snapshots
model_configs
exports
```

## Collection: friction_events

Stores manually logged friction events.

Example:

```json
{
  "_id": "ObjectId",
  "schema_version": 1,
  "timestamp_start": "2026-06-01T09:15:00Z",
  "timestamp_end": "2026-06-01T09:35:00Z",
  "workflow_stage": "test",
  "friction_layer": "technical",
  "friction_type": "failed_feedback",
  "severity_self": 4,
  "cognitive_load_self": 3,
  "emotion_valence": -1,
  "time_lost_minutes": 20,
  "resume_time_minutes": 8,
  "recovery_minutes": 0,
  "interruption_count": 1,
  "goal_id": "ObjectId",
  "session_id": "ObjectId",
  "tags": ["ci", "flaky-test"],
  "notes": "CI failed with a test that passed locally.",
  "source": "manual",
  "created_at": "2026-06-01T09:36:00Z",
  "updated_at": "2026-06-01T09:36:00Z"
}
```

Required fields:

```text
schema_version
timestamp_start
workflow_stage
friction_layer
friction_type
severity_self
cognitive_load_self
emotion_valence
time_lost_minutes
resume_time_minutes
interruption_count
source
created_at
updated_at
```

Optional fields:

```text
timestamp_end
recovery_minutes
goal_id
session_id
tags
notes
```

Recommended indexes:

```text
timestamp_start
workflow_stage
friction_layer
friction_type
goal_id
session_id
created_at
```

Compound indexes:

```text
{ "timestamp_start": -1, "workflow_stage": 1 }
{ "timestamp_start": -1, "friction_layer": 1 }
{ "timestamp_start": -1, "friction_type": 1 }
```

## Collection: work_sessions

Stores bounded work sessions.

Example:

```json
{
  "_id": "ObjectId",
  "schema_version": 1,
  "title": "Morning feature work",
  "started_at": "2026-06-01T08:30:00Z",
  "ended_at": "2026-06-01T11:30:00Z",
  "goal_ids": ["ObjectId"],
  "notes": "Worked on dashboard filtering.",
  "created_at": "2026-06-01T08:30:00Z",
  "updated_at": "2026-06-01T11:30:00Z"
}
```

Recommended indexes:

```text
started_at
ended_at
created_at
```

## Collection: work_goals

Stores user-defined work goals.

Example:

```json
{
  "_id": "ObjectId",
  "schema_version": 1,
  "title": "Implement friction event filters",
  "description": "Add filtering by workflow stage, layer, and type.",
  "status": "active",
  "created_at": "2026-06-01T08:00:00Z",
  "updated_at": "2026-06-01T08:00:00Z"
}
```

Allowed statuses:

```text
active
completed
deferred
abandoned
```

Recommended indexes:

```text
status
created_at
updated_at
```

## Collection: score_snapshots

Stores computed score outputs.

Example:

```json
{
  "_id": "ObjectId",
  "schema_version": 1,
  "model_version": "model-0.1",
  "model_config_id": "ObjectId",
  "period_start": "2026-06-01T00:00:00Z",
  "period_end": "2026-06-07T23:59:59Z",
  "score_type": "weekly",
  "scores": {
    "cla": 37.4,
    "fci": 21.8,
    "sdc": 0.42
  },
  "event_scores": [
    {
      "event_id": "ObjectId",
      "fcs": 14.6
    }
  ],
  "top_contributors": [
    {
      "event_id": "ObjectId",
      "reason": "high severity and long resume time"
    }
  ],
  "created_at": "2026-06-07T23:59:59Z",
  "updated_at": "2026-06-07T23:59:59Z"
}
```

Recommended indexes:

```text
period_start
period_end
score_type
model_version
created_at
```

Compound index:

```text
{ "period_start": 1, "period_end": 1, "score_type": 1 }
```

## Collection: model_configs

Stores scoring model parameters.

Example:

```json
{
  "_id": "ObjectId",
  "schema_version": 1,
  "model_version": "model-0.1",
  "name": "Default model",
  "parameters": {
    "cla_decay": 0.85,
    "severity_multiplier": 1.2,
    "cognitive_load_multiplier": 1.5,
    "interruption_multiplier": 2.0,
    "recovery_multiplier": 0.3,
    "fci_half_life_minutes": 90
  },
  "is_default": true,
  "created_at": "2026-06-01T00:00:00Z",
  "updated_at": "2026-06-01T00:00:00Z"
}
```

Recommended indexes:

```text
model_version
is_default
created_at
```

## Collection: exports

Stores export metadata.

Example:

```json
{
  "_id": "ObjectId",
  "schema_version": 1,
  "export_type": "json",
  "status": "completed",
  "period_start": "2026-06-01T00:00:00Z",
  "period_end": "2026-06-07T23:59:59Z",
  "file_path": "./exports/logarift-2026-06-01-2026-06-07.json",
  "created_at": "2026-06-07T23:59:59Z",
  "updated_at": "2026-06-07T23:59:59Z"
}
```

Allowed export statuses:

```text
pending
completed
failed
```

Recommended indexes:

```text
created_at
export_type
status
```

## Validation

initial release may use application-level validation in Go.

MongoDB JSON schema validation may be added later for stronger guarantees.

## Schema Versioning

Documents must include `schema_version`.

Rules:

- new fields should be optional when possible
- breaking changes require schema_version increment
- readers should tolerate unknown fields
- migration scripts may be added later if needed

## Data Retention

initial release keeps data until the user deletes it.

Future versions may include local retention settings.

## Quick Logging Extension

Quick friction events add an observed/inferred/canonical structure while keeping the original top-level fields for backwards compatibility.

Additional fields on `friction_events`:

```text
input_mode
observed
inference
canonical
```

`input_mode` values:

```text
quick
advanced
```

Example quick event extension:

```json
{
  "input_mode": "quick",
  "observed": {
    "occurred_at": "2026-06-04T19:26:00Z",
    "friction_level": "orange",
    "notes_markdown": "CI failed again after 20 min with an unclear timeout.",
    "plain_text": "CI failed again after 20 min with an unclear timeout.",
    "links": []
  },
  "inference": {
    "engine_version": "rules-0.1",
    "engine_type": "rules",
    "created_at": "2026-06-04T19:26:05Z",
    "fields": {
      "workflow_stage": {
        "value": "test",
        "confidence": 0.88,
        "source": "rules",
        "explanation": "Matched testing, CI, pipeline, or flaky-test language."
      }
    }
  },
  "canonical": {
    "workflow_stage": "test",
    "friction_layer": "technical",
    "friction_type": "failed_feedback",
    "severity_self": 4,
    "cognitive_load_self": 4,
    "emotion_valence": -1,
    "time_lost_minutes": 20,
    "resume_time_minutes": 8,
    "recovery_minutes": 0,
    "interruption_count": 0,
    "tags": ["ci", "timeout"]
  }
}
```

The canonical values are also copied to the existing top-level fields such as `workflow_stage`, `friction_layer`, `friction_type`, `severity_self`, `time_lost_minutes`, and `tags`. This allows existing list filters, dashboards, and math-engine scoring to work without a migration.

## Local Uploaded Images

Screenshots and other rich-note images are stored as local files under `LOGARIFT_UPLOAD_DIR` and served through `/uploads/{filename}`. Event notes store image URLs instead of embedding binary image bytes in MongoDB. This keeps `friction_events` documents small while preserving local-first behavior.
