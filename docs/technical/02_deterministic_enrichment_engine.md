# Deterministic Enrichment Engine

## Summary

The backend now includes a local deterministic enrichment engine used by the quick friction logging endpoint.

The engine lives in:

```text
backend/internal/enrichment
```

The API integration lives in:

```text
backend/internal/friction.Service.CreateQuick
POST /api/v1/friction-events/quick
```

## Runtime flow

```text
React quick composer
  -> optional POST /api/v1/uploads for pasted or attached screenshots
  -> POST /api/v1/friction-events/quick
  -> friction service validates three observed fields
  -> enrichment engine creates observed/inference/canonical event shape
  -> canonical fields are copied to top-level event fields
  -> MongoDB stores the event
  -> dashboard and math engine continue using canonical top-level fields
```

## Input contract

```json
{
  "occurred_at": "2026-06-04T19:26:00Z",
  "friction_level": "orange",
  "notes_markdown": "CI failed again after 20 min with an unclear timeout.",
  "links": []
}
```

Only the first three fields are required by the UI. `links` is optional; the engine also extracts links from notes.

## Output contract

The service returns a normal `friction_event` document with additional metadata:

```text
input_mode
observed
inference
canonical
```

Canonical fields are duplicated to the existing top-level fields for backwards compatibility with:

- list filters
- dashboard grouping
- score calculation
- existing CRUD endpoint behavior

## Classification strategy

The rule engine uses simple keyword matching and priority rules.

Examples:

```text
CI + failed + timeout      -> test / technical / failed_feedback
PR + waiting + approval    -> code_review / social_process / waiting_for_review
docs + missing             -> documentation / cognitive / missing_documentation
interrupted + meeting      -> coordination / environmental / interruption
```

The engine returns confidence values for inferred fields. Confidence is not used to change the initial release math score yet; it is displayed as data-quality metadata in the UI.

## Why this is separate from the math engine

The C++ math engine remains deterministic score calculation over canonical fields.

The enrichment engine owns uncertain interpretation from human notes.

This keeps the architecture clear:

```text
enrichment = classify and estimate fields from notes
math       = calculate scores from canonical event data
```

## Extension points

Future engines should implement the same conceptual contract:

```text
input: occurred_at + friction_level + notes + links + attachments
output: observed + inference + canonical fields
```

Possible future implementations:

- rules-0.2 with better keyword dictionaries
- local LLM adapter
- local ML classifier trained from corrections
- URL metadata enrichers for GitHub Actions, Jira, documentation, or CI systems
- screenshot attachment metadata and later local OCR

Any future inference engine must preserve:

- local-first operation
- explicit engine version
- field-level confidence
- fallback to deterministic rules
- no hidden telemetry

## Upload and Rich Notes Boundary

The frontend rich notes editor uploads screenshots to the backend through `POST /api/v1/uploads`. The backend stores accepted images under `LOGARIFT_UPLOAD_DIR` and serves them from `/uploads/{filename}`. The quick friction event stores note HTML containing image URLs rather than storing image bytes in MongoDB.

The deterministic enrichment engine currently uses note text and links. It does not inspect image pixels. Future local OCR or image understanding should be implemented as a separate optional enrichment extension.
