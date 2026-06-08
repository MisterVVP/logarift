# Quick Logging and Deterministic Enrichment

## Purpose

The first UI exposed too many fields for a friction logging product. A developer should not have to classify workflow stage, layer, type, severity, cognitive load, resume time, interruption count, tags, goal, and session while already experiencing friction.

The new default logging model is:

```text
Developer records the friction.
Logarift classifies and estimates the structured fields.
Developer may correct the result later when useful.
```

## Default Quick Logging Fields

The default UI should ask for at most three fields:

```text
occurred_at
friction_level
notes_markdown
```

Field meaning:

- `occurred_at`: when the friction happened; defaults to now.
- `friction_level`: a color-coded frustration/severity input.
- `notes_markdown`: a rich note field that can contain formatted text, links, and uploaded/pasted screenshot references.

## Friction Levels

The UI maps color choices to canonical scoring fields.

```text
green  = papercut / small annoyance
yellow = annoying slowdown
orange = disruptive friction
red    = blocker or severe frustration
```

Initial deterministic mapping:

```text
green:
  severity_self = 1
  cognitive_load_self = 1
  emotion_valence = 0
  default_time_lost_minutes = 2
  default_resume_time_minutes = 0

yellow:
  severity_self = 2
  cognitive_load_self = 2
  emotion_valence = -1
  default_time_lost_minutes = 10
  default_resume_time_minutes = 2

orange:
  severity_self = 4
  cognitive_load_self = 4
  emotion_valence = -1
  default_time_lost_minutes = 30
  default_resume_time_minutes = 8

red:
  severity_self = 5
  cognitive_load_self = 5
  emotion_valence = -2
  default_time_lost_minutes = 60
  default_resume_time_minutes = 15
```

Explicit durations in notes override the default time-loss estimate.

## Observed, Inferred, Canonical Shape

Quick events are stored with three layers of meaning.

```text
observed  = what the developer explicitly entered
inference = deterministic local interpretation with confidence metadata
canonical = current analytics/math-compatible fields
```

Example:

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
    "fields": {
      "workflow_stage": {
        "value": "test",
        "confidence": 0.88,
        "source": "rules"
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
    "interruption_count": 0
  }
}
```

For backwards compatibility, canonical values are also copied to the existing top-level event fields used by dashboards, filters, and the C++ math engine.

## Deterministic Rule Engine

The first enrichment engine is local and deterministic.

It performs:

- friction-level to score mapping
- duration parsing from notes
- URL extraction
- workflow stage classification
- friction layer classification
- friction type classification
- resume-time estimation
- interruption-count estimation
- tag extraction
- confidence metadata generation

The current engine version is:

```text
rules-0.1
```

The engine does not call cloud services and does not require a local LLM.

## API

Quick logging endpoint:

```text
POST /api/v1/friction-events/quick
```

Request:

```json
{
  "occurred_at": "2026-06-04T19:26:00Z",
  "friction_level": "orange",
  "notes_markdown": "CI failed again after 20 min with an unclear timeout."
}
```

Response:

```json
{
  "event": {
    "input_mode": "quick",
    "workflow_stage": "test",
    "friction_layer": "technical",
    "friction_type": "failed_feedback",
    "observed": {},
    "inference": {},
    "canonical": {}
  }
}
```

Screenshot and image upload endpoint:

```text
POST /api/v1/uploads
GET /uploads/{filename}
```

Images are stored locally in the backend upload directory and referenced from notes as normal image URLs. Large binary content is not embedded directly into MongoDB event documents.

The old full endpoint remains available for advanced/manual correction workflows:

```text
POST /api/v1/friction-events
```

## Future Local LLM Adapter

A future optional adapter may use a local LLM to improve ontology extraction.

Design constraints:

- disabled by default
- local-only
- no hidden telemetry
- no cloud LLM requirement
- structured JSON output only
- deterministic settings where possible
- fallback to deterministic rules when output is invalid
- confidence and explanation required for every inferred field

Possible local-first runtime options include tools such as Ollama or other local model runtimes, but this initial release iteration does not implement them.

## Future Local ML Classifier

A later local ML classifier may learn from user corrections.

Design constraints:

- trained on local data only unless user explicitly exports data
- optional
- explainable enough for correction workflows
- versioned model metadata
- deterministic fallback rules remain available

Possible implementation options include ONNX Runtime or a small embedded classifier. This initial release iteration only documents the extension point.

## Product Rule

The default product rule is:

```text
Ask before save only what the developer knows immediately.
Infer after save what the system can reasonably classify.
Allow correction later, but never make correction mandatory.
```

## UX Layout Rule

The default screen must show only the quick friction composer and recent logs. Optional context such as goals and sessions is hidden behind a modal. Analytics live on a separate Dashboard tab. This preserves the product rule that logging friction must not create new friction.

## Tooltip Rule

Complex concepts such as avg cognitive load, avg inference confidence, CLA, FCI, and SDC must have visible tooltip affordances. The tooltip should explain what the value means without requiring the user to read product docs.
