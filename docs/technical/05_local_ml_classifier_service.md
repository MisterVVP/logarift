# Local ML Classifier Service

## Summary

The Local ML Classifier is a future optional local-only service that can learn from user corrections and produce fast structured field predictions for quick friction events.

It complements the deterministic rules engine and the Local LLM Adapter. It does not replace backend validation, deterministic fallback, or the C++ math engine.

## Goals

The classifier should:

- train only from local Logarift data unless a user explicitly exports data
- run inference locally with bounded CPU and memory usage
- provide fast predictions for ontology fields and tags
- expose model metadata and training provenance
- support deterministic fallback rules for every field
- support explainable correction workflows through feature hints and confidence values
- avoid requiring a GPU or large model download

## Non-goals

The classifier must not:

- send local notes or labels to cloud training services
- rank developers or infer individual productivity
- generate coaching or recommendations
- update canonical events without backend validation
- train automatically without explicit local configuration
- require the Local LLM Adapter to be enabled
- change math-engine scoring formulas

## Relationship to other enrichment components

```text
Observed quick event input
  -> deterministic rules baseline
  -> optional Local ML Classifier prediction
  -> optional Local LLM Adapter prediction
  -> backend merge and validation policy
  -> canonical event fields
  -> C++ math engine score calculation
```

Recommended responsibility split:

```text
rules engine       = safe baseline, duration parsing, URL extraction, simple keywords
ML classifier      = fast learned classification from local corrections
LLM adapter        = optional deeper interpretation of ambiguous notes
math engine        = deterministic scoring over canonical fields
```

The backend should be able to run with only rules, rules plus ML, rules plus LLM, or all three.

## Service boundary

The classifier should run as a separate local HTTP service.

Recommended service name:

```text
ml-classifier
```

Recommended default port:

```text
8092
```

Reasons for a separate service:

- isolates model loading and training from the API process
- allows independent CPU and memory limits
- keeps backend dependencies small
- permits use of ML-specific runtimes such as ONNX Runtime, scikit-learn-compatible exports, or a small native inference library
- keeps model file ownership and versioning explicit

## First-version model approach

The first classifier should be intentionally small and explainable.

Recommended initial model family:

```text
text features + linear classifier or gradient-boosted shallow classifier
```

Recommended features:

- normalized word and character n-grams
- deterministic rule hits as binary features
- friction level
- explicit duration-present flag
- URL host categories such as local Git, CI, docs, issue tracker, or unknown
- attachment-present flag
- recent user correction frequencies by field value

Recommended prediction targets:

```text
workflow_stage
friction_layer
friction_type
tags
resume_time_minutes bucket
interruption_count bucket
```

Numeric canonical fields that directly affect scoring should remain conservative. The classifier may suggest buckets, but the backend should convert or reject them through explicit policy.

## Training data source

Training examples should come from local correction history, not hidden telemetry.

A training row should include:

```text
event_id
created_at
corrected_at
observed plain text features
friction_level
baseline rule output
previous inferred value
user-corrected canonical value
field_name
label
schema_version
```

The backend can expose local correction examples to the classifier through a controlled export endpoint or write a local training dataset file. The classifier should not read MongoDB directly in the first version.

## Model artifact layout

Suggested local model directory:

```text
data/models/ml-classifier/
  active/
    model.onnx
    metadata.json
    vocabulary.json
    labels.json
  candidates/
    2026-06-05T120000Z/
      model.onnx
      metadata.json
      evaluation.json
```

Suggested `metadata.json`:

```json
{
  "model_id": "local-ml-classifier-2026-06-05T120000Z",
  "model_version": "ml-classifier-0.1",
  "schema_version": "ml-classifier-metadata-v1",
  "trained_at": "2026-06-05T12:00:00Z",
  "training_example_count": 184,
  "training_event_count": 92,
  "features_version": "text-features-0.1",
  "labels_version": "ontology-0.1",
  "metrics": {
    "workflow_stage_macro_f1": 0.71,
    "friction_layer_macro_f1": 0.68,
    "friction_type_macro_f1": 0.62
  },
  "activation_status": "candidate"
}
```

The active model must be switchable atomically so inference never reads a partially written artifact.

## Inference request contract

Endpoint:

```text
POST /v1/predict/friction-event
```

Request:

```json
{
  "request_id": "local-generated-request-id",
  "schema_version": "ml-classifier-predict-request-v1",
  "occurred_at": "2026-06-04T19:26:00Z",
  "observed": {
    "friction_level": "orange",
    "plain_text": "CI failed again after 20 min with an unclear timeout.",
    "links": [],
    "attachment_metadata": []
  },
  "deterministic_baseline": {
    "workflow_stage": "test",
    "friction_layer": "technical",
    "friction_type": "failed_feedback",
    "tags": ["ci", "timeout"]
  },
  "allowed_values": {
    "workflow_stage": ["planning", "local_development", "build", "test", "code_review", "merge", "deployment", "operation", "debugging", "documentation", "coordination", "learning"],
    "friction_layer": ["technical", "temporal", "cognitive", "social_process", "motivational", "environmental"],
    "friction_type": ["slow_feedback", "failed_feedback", "unclear_error", "missing_documentation", "ambiguous_ownership", "interruption", "waiting_for_review", "waiting_for_ci", "context_switch", "rework", "tooling_failure", "environment_setup", "coordination_overhead", "decision_blocked"]
  }
}
```

The request should avoid rich markdown unless a feature extractor specifically needs it. Plain text is preferred for privacy and predictable feature extraction.

## Inference response contract

Response:

```json
{
  "schema_version": "ml-classifier-predict-response-v1",
  "request_id": "local-generated-request-id",
  "service_version": "ml-classifier-service-0.1",
  "model_id": "local-ml-classifier-2026-06-05T120000Z",
  "model_version": "ml-classifier-0.1",
  "features_version": "text-features-0.1",
  "duration_ms": 18,
  "fields": {
    "workflow_stage": {
      "value": "test",
      "confidence": 0.87,
      "source": "local_ml",
      "evidence": ["token:ci", "token:failed", "token:timeout", "rule:test"]
    },
    "friction_layer": {
      "value": "technical",
      "confidence": 0.81,
      "source": "local_ml",
      "evidence": ["rule:technical", "token:ci"]
    },
    "friction_type": {
      "value": "failed_feedback",
      "confidence": 0.76,
      "source": "local_ml",
      "evidence": ["token:failed", "token:timeout"]
    },
    "tags": {
      "value": ["ci", "timeout"],
      "confidence": 0.79,
      "source": "local_ml",
      "evidence": ["token:ci", "token:timeout"]
    }
  },
  "warnings": []
}
```

Response rules:

- values must be from backend-supplied allowed values
- confidence must be calibrated as well as possible and always bounded to `0.0` through `1.0`
- evidence must be short feature identifiers, not raw note excerpts
- unknown labels should be returned when confidence is low
- the backend remains the source of truth for accepted canonical fields

## Training API

Training should be explicit and local.

Recommended endpoints:

```text
POST /v1/train/jobs
GET /v1/train/jobs/{job_id}
POST /v1/models/{model_id}/activate
GET /v1/models
GET /v1/models/active
```

Training job request:

```json
{
  "schema_version": "ml-classifier-train-request-v1",
  "training_dataset_path": "/data/training/corrections-2026-06-05.jsonl",
  "min_examples_per_label": 5,
  "activate_if_metrics_pass": false
}
```

The default should create a candidate model, not activate it automatically.

## Activation policy

A candidate model may be activated only when all configured gates pass.

Suggested initial gates:

```text
minimum total examples: 50
minimum examples per predicted enum label: 5
workflow_stage macro F1: >= 0.60
friction_layer macro F1: >= 0.60
friction_type macro F1: >= 0.55
no schema mismatch
no unknown ontology values
```

If gates fail, the candidate remains available for inspection but must not be used for inference by default.

## Backend merge policy

The backend should merge classifier predictions field-by-field.

Recommended initial gates:

```text
workflow_stage: accept local_ml when confidence >= 0.75 and value is allowed
friction_layer: accept local_ml when confidence >= 0.75 and value is allowed
friction_type: accept local_ml when confidence >= 0.80 and value is allowed
tags: union deterministic tags and accepted local_ml tags, then normalize and cap
resume_time_minutes bucket: accept only when confidence >= 0.85, otherwise keep deterministic estimate
interruption_count bucket: accept only when confidence >= 0.85, otherwise keep deterministic estimate
```

When both Local ML and Local LLM are enabled, a conservative initial policy is:

```text
1. observed explicit values win
2. deterministic duration parsing wins for time_lost_minutes
3. if ML and LLM agree on an allowed enum and both pass confidence gates, accept that value
4. if only one optional engine passes confidence gates, accept it only if it beats the deterministic baseline gate
5. otherwise keep deterministic baseline
```

## Configuration

Suggested backend variables:

```text
LOGARIFT_ML_CLASSIFIER_ENABLED=false
LOGARIFT_ML_CLASSIFIER_URL=http://localhost:8092
LOGARIFT_ML_CLASSIFIER_TIMEOUT_MS=500
LOGARIFT_ML_CLASSIFIER_MIN_CONFIDENCE=0.75
```

Suggested classifier variables:

```text
LOGARIFT_ML_CLASSIFIER_PORT=8092
LOGARIFT_ML_MODEL_DIR=/data/models/ml-classifier
LOGARIFT_ML_TRAINING_DIR=/data/training
LOGARIFT_ML_MAX_INPUT_CHARS=6000
LOGARIFT_ML_LOG_FEATURES=false
LOGARIFT_ML_AUTO_TRAIN=false
LOGARIFT_ML_AUTO_ACTIVATE=false
```

The classifier must be disabled by default in the backend until explicitly configured.

## Reliability requirements

The classifier should be fast and predictable.

Backend requirements:

- short inference timeout
- strict response size limit
- deterministic fallback on timeout or invalid response
- no event-save dependency on classifier availability
- circuit-breaker or cooldown after repeated failures

Classifier service requirements:

- load active model at startup or report not-ready
- expose clear readiness when no active model exists
- avoid partial artifact reads through atomic activation
- keep training jobs separate from inference request handling
- rate-limit or serialize local training jobs to avoid starving inference

## Security and privacy requirements

The classifier must preserve local-first privacy guarantees:

- do not send training data or metrics externally
- do not log raw notes by default
- store artifacts in a local documented directory
- avoid reading arbitrary filesystem paths from API requests unless the path is under an allowlisted training directory
- validate all dataset paths and model IDs
- never execute model files as code
- reject schema-incompatible model artifacts

## Observability

Inference logs should include:

```text
request_id
status
service_version
model_id
duration_ms
accepted_prediction_count
warning_count
error_code
```

Training logs should include:

```text
job_id
status
training_example_count
model_id
metrics_summary
duration_ms
error_code
```

Logs should not include raw notes or full feature vectors by default.

## Health endpoints

Recommended endpoints:

```text
GET /health/live
GET /health/ready
GET /v1/models/active
POST /v1/predict/friction-event
POST /v1/train/jobs
GET /v1/train/jobs/{job_id}
POST /v1/models/{model_id}/activate
```

Readiness should fail when no active model is loaded if classifier inference is enabled. A separate candidate-training mode may remain live even when no active model exists.

## Testing strategy

Implementation should include:

- feature extraction unit tests
- schema validation tests
- model metadata validation tests
- deterministic fixture inference tests
- backend merge-policy tests
- timeout and invalid response fallback tests
- training job tests with tiny local fixtures
- artifact activation atomicity tests
- Docker Compose smoke test with classifier disabled

Default tests must not require network access or downloading pretrained models.

## Rollout plan

Recommended phases:

1. Add correction-history export shape and local dataset generation.
2. Add classifier service skeleton with health and fake model mode.
3. Add feature extraction and schema validation.
4. Add backend client, disabled-by-default configuration, and merge tests.
5. Add first small local classifier and candidate artifact layout.
6. Add explicit training job workflow.
7. Add model activation gates and active-model metadata display.
8. Add optional UI indicators for `local_ml` source, confidence, and correction feedback.

## Open decisions

Before implementation, decide:

- exact model runtime and artifact format
- whether the service is written in Go, Python, or C++
- correction-history storage schema
- minimum local data threshold before enabling training
- label taxonomy versioning strategy
- how model metrics are shown to the user
- whether model activation can happen from the UI or only through local admin/API calls
