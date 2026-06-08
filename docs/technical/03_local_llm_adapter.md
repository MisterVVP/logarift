# Local LLM Adapter Service

## Summary

The Local LLM Adapter is an optional local-only service that can improve quick friction event enrichment by converting notes, links, and attachment metadata into structured ontology fields.

It must sit behind the existing enrichment boundary and must never replace deterministic rules as the safe fallback path.

The adapter is not part of the C++ math engine. It interprets unstructured inputs into canonical event candidates; the math engine continues to calculate scores from canonical structured fields only.

## Goals

The adapter should:

- run entirely on the user's machine or private local network
- be disabled by default until explicitly configured
- call only approved local model runtimes
- return strict structured JSON
- include field-level confidence and short explanations
- preserve deterministic rule fallback for invalid, slow, unavailable, or low-confidence responses
- avoid hidden telemetry, cloud calls, and prompt/content logging by default
- make model version, prompt version, and adapter version visible in stored inference metadata

## Non-goals

The adapter must not:

- call hosted LLM APIs in initial release local mode
- generate productivity recommendations or behavioral coaching
- score developer performance
- modify the C++ math formulas
- write directly to MongoDB
- bypass backend validation or ontology validation
- require a GPU to run the default local development stack
- make event saving depend on LLM availability

## Runtime topology

```text
React quick composer
  -> Go backend API
  -> deterministic enrichment engine
       -> optional Local LLM Adapter client
            -> Local LLM Adapter service
                 -> local model runtime such as Ollama-compatible HTTP runtime
       -> validation, confidence gating, and deterministic fallback
  -> MongoDB stores observed/inference/canonical event
  -> C++ math engine scores canonical fields later
```

The Go backend owns orchestration and validation. The adapter owns prompt construction, local runtime calls, response repair within strict limits, and normalized JSON output.

## Service boundary

The adapter should be implemented as a separate local HTTP service instead of embedding model-runtime code into the Go backend.

Reasons:

- isolates model runtime failures from the API process
- allows independent resource limits and timeouts
- avoids coupling backend releases to model runtime clients
- allows the backend to treat LLM enrichment as an optional dependency
- keeps Docker Compose service ownership explicit
- makes it easier to support multiple local runtimes later

Recommended service name:

```text
llm-adapter
```

Recommended default port:

```text
8091
```

## Resolved v1 decisions

The first implementation should use these decisions:

```text
runtime protocol: Ollama-compatible local HTTP API
runtime endpoint: POST /api/chat with stream=false and structured JSON output
adapter language: Go
preferred first model: qwen3.6 through Ollama
smaller fallback model: qwen3:8b when qwen3.6 is too large for the machine
```

"Runtime protocol" means the wire contract the adapter uses to call the local model runtime. For v1, the adapter should call a local Ollama server over HTTP at `http://localhost:11434/api`, not embed a model runtime directly and not call an external hosted LLM API. `POST /api/chat` is preferred over raw generation because the request naturally separates system instructions from user event text and matches Ollama's Qwen3.6 examples.

Go is the default adapter implementation language because the adapter is mostly HTTP, schema validation, prompt templating, timeout handling, and JSON normalization. Python can still be used later for offline prompt experiments or evaluation scripts, but the production local service should be a small Go binary unless implementation evidence shows that Python materially reduces risk.

The ontology enum list supplied to the adapter must be generated from the backend ontology source of truth and should initially be:

```text
workflow_stage:
  planning
  local_development
  build
  test
  code_review
  merge
  deployment
  operation
  debugging
  documentation
  coordination
  learning

friction_layer:
  technical
  temporal
  cognitive
  social_process
  motivational
  environmental

friction_type:
  slow_feedback
  failed_feedback
  unclear_error
  missing_documentation
  ambiguous_ownership
  interruption
  waiting_for_review
  waiting_for_ci
  context_switch
  rework
  tooling_failure
  environment_setup
  coordination_overhead
  decision_blocked
```

The adapter must not invent values outside those lists. If the model cannot choose safely, the adapter should omit the field or return a rejected suggestion; the backend then uses deterministic rules.

## Backend integration point

The existing quick logging path remains the integration point:

```text
POST /api/v1/friction-events/quick
```

The backend should continue to:

1. validate observed quick input
2. call the Local LLM Adapter only when enabled and safe to reach
3. validate the adapter response against the ontology and schema
4. run deterministic rules after the adapter response is available
5. use deterministic rules to correct, fill, or replace any missing, invalid, low-confidence, or contradictory adapter field
6. merge accepted LLM suggestions and deterministic fields into inference/canonical output using confidence gates
7. persist accepted and rejected model suggestions as MongoDB inference metadata
8. store the final observed/inference/canonical shape

The adapter should not expose a user-facing browser API in initial release. Browser traffic continues to go through the Go backend.

## Request contract

Endpoint:

```text
POST /v1/enrich/friction-event
```

Request:

```json
{
  "request_id": "local-generated-request-id",
  "schema_version": "llm-adapter-request-v1",
  "occurred_at": "2026-06-04T19:26:00Z",
  "observed": {
    "friction_level": "orange",
    "notes_markdown": "CI failed again after 20 min with an unclear timeout.",
    "plain_text": "CI failed again after 20 min with an unclear timeout.",
    "links": [],
    "attachment_metadata": []
  },
  "deterministic_baseline": {
    "workflow_stage": "test",
    "friction_layer": "technical",
    "friction_type": "failed_feedback",
    "severity_self": 4,
    "cognitive_load_self": 4,
    "emotion_valence": -1,
    "time_lost_minutes": 20,
    "resume_time_minutes": 8,
    "interruption_count": 0,
    "tags": ["ci", "timeout"]
  },
  "allowed_values": {
    "workflow_stage": ["planning", "local_development", "build", "test", "code_review", "merge", "deployment", "operation", "debugging", "documentation", "coordination", "learning"],
    "friction_layer": ["technical", "temporal", "cognitive", "social_process", "motivational", "environmental"],
    "friction_type": ["slow_feedback", "failed_feedback", "unclear_error", "missing_documentation", "ambiguous_ownership", "interruption", "waiting_for_review", "waiting_for_ci", "context_switch", "rework", "tooling_failure", "environment_setup", "coordination_overhead", "decision_blocked"]
  }
}
```

Request rules:

- `plain_text` should be sanitized text extracted from rich notes.
- `notes_markdown` may be omitted if prompt privacy mode is set to text-only.
- `attachment_metadata` should include only local metadata in the first version, not raw image bytes.
- `allowed_values` must be supplied by the backend so the adapter cannot invent ontology values.
- `deterministic_baseline` lets the model improve or confirm existing fields instead of starting from an unconstrained blank prompt.

## Response contract

Response:

```json
{
  "schema_version": "llm-adapter-response-v1",
  "request_id": "local-generated-request-id",
  "adapter_version": "llm-adapter-0.1",
  "model_runtime": "ollama-compatible",
  "model_name": "configured-local-model-name",
  "model_digest": "optional-local-model-digest",
  "prompt_version": "friction-enrichment-prompt-0.1",
  "duration_ms": 842,
  "fields": {
    "workflow_stage": {
      "value": "test",
      "confidence": 0.9,
      "source": "local_llm",
      "explanation": "The note describes CI failure and timeout during validation."
    },
    "friction_layer": {
      "value": "technical",
      "confidence": 0.86,
      "source": "local_llm",
      "explanation": "The blocker is caused by test infrastructure behavior."
    },
    "friction_type": {
      "value": "failed_feedback",
      "confidence": 0.82,
      "source": "local_llm",
      "explanation": "The feedback loop failed with an unclear timeout."
    },
    "time_lost_minutes": {
      "value": 20,
      "confidence": 0.95,
      "source": "observed_text",
      "explanation": "The note explicitly mentions 20 minutes."
    },
    "tags": {
      "value": ["ci", "timeout"],
      "confidence": 0.8,
      "source": "local_llm",
      "explanation": "The note mentions CI and timeout."
    }
  },
  "warnings": []
}
```

Response rules:

- all ontology values must be strings from `allowed_values`
- numeric fields must stay within backend-defined ranges
- confidence must be in the inclusive range `0.0` to `1.0`
- explanations must be short, local, and non-prescriptive
- the backend must reject unknown fields unless explicitly allowlisted
- the backend must treat warnings as metadata, not as successful validation

## Merge policy

The backend should merge adapter fields field-by-field, never all-or-nothing unless the whole response is invalid JSON.

Recommended initial gates:

```text
workflow_stage: accept local_llm when confidence >= 0.70 and value is allowed
friction_layer: accept local_llm when confidence >= 0.70 and value is allowed
friction_type: accept local_llm when confidence >= 0.75 and value is allowed
time_lost_minutes: prefer explicit observed duration; otherwise accept local_llm when confidence >= 0.85
resume_time_minutes: keep deterministic estimate unless local_llm confidence >= 0.85
interruption_count: keep deterministic estimate unless local_llm confidence >= 0.85
tags: union deterministic tags and accepted local_llm tags, then normalize and cap
```

If a field is rejected, the deterministic output remains canonical for that field. Rejected model suggestions should still be persisted in MongoDB as inference metadata so the user and future correction workflows can inspect what happened. Rejected metadata must include the suggested value, confidence, rejection reason, adapter version, model name, and prompt version, but it must not duplicate raw note text beyond what is already stored in the event.

## Prompting constraints

Prompts should be built from templates checked into the adapter service code. Runtime prompt construction should inject only:

- event text
- allowed ontology values
- deterministic baseline
- requested JSON schema
- short task instructions

Prompt rules:

- use low temperature or deterministic runtime settings when available
- instruct the model to return JSON only
- instruct the model to choose `unknown` instead of inventing categories
- never ask for productivity judgments, performance ratings, or advice
- never include unrelated local files, environment variables, secrets, or repository content
- cap prompt size and truncate notes safely with clear metadata

## Prompt size and truncation

Prompt size still matters for local models even when the hardware is strong. Larger prompts increase latency, memory pressure, and the probability that the model ignores the schema. The adapter should therefore enforce explicit limits instead of relying only on the model context window.

Initial limits:

```text
maximum sanitized note input: 12,000 Unicode characters
maximum complete prompt target: 8,192 tokens when the runtime exposes token estimates
reserved output budget: at least 1,024 tokens
default truncation: head_tail
```

`head_tail` keeps the start and end of long notes and replaces the middle with a clear truncation marker. This preserves the initial problem statement and the latest error/status details, which are usually the most useful parts of a friction note. The adapter should include `truncated=true`, original character count, retained character count, and strategy in response metadata.

## Configuration

Suggested backend variables:

```text
LOGARIFT_LLM_ADAPTER_ENABLED=false
LOGARIFT_LLM_ADAPTER_URL=http://localhost:8091
LOGARIFT_LLM_ADAPTER_TIMEOUT_MS=1500
LOGARIFT_LLM_ADAPTER_MIN_CONFIDENCE=0.70
LOGARIFT_LLM_ADAPTER_PROMPT_PRIVACY_MODE=text_only
LOGARIFT_LLM_ADAPTER_ALLOW_REMOTE_RUNTIME=false
```

Suggested adapter variables:

```text
LOGARIFT_LLM_ADAPTER_PORT=8091
LOGARIFT_LLM_RUNTIME_URL=http://localhost:11434
LOGARIFT_LLM_ADAPTER_ALLOW_REMOTE_RUNTIME=false
LOGARIFT_LLM_MODEL=qwen3.6
LOGARIFT_LLM_REQUEST_TIMEOUT_MS=1200
LOGARIFT_LLM_MAX_INPUT_CHARS=12000
LOGARIFT_LLM_MAX_PROMPT_TOKENS=8192
LOGARIFT_LLM_TRUNCATION_STRATEGY=head_tail
LOGARIFT_LLM_LOG_PROMPTS=false
LOGARIFT_LLM_LOG_RESPONSES=false
```

Default behavior must be disabled unless `LOGARIFT_LLM_ADAPTER_ENABLED=true` is explicitly set for the backend.

## Reliability requirements

The backend should enforce:

- small connection timeout
- total request deadline
- strict JSON decoder size limit
- circuit-breaker or short cooldown after repeated adapter failures
- deterministic fallback on timeout, connection failure, invalid JSON, invalid schema, or validation failure
- no retries for a single quick-save request unless the retry can complete inside the request deadline

The quick logging endpoint should save the event even when the LLM adapter is unavailable.

## Security and privacy requirements

The adapter must preserve local-first privacy guarantees:

- bind to localhost by default for direct local runs
- expose only local Docker network ports by default in Compose
- do not collect telemetry
- do not log prompts or model responses by default
- redact request IDs and metadata rather than note bodies in operational logs
- reject runtime URLs outside localhost, loopback, or configured private Docker networks by default; allow them only when `LOGARIFT_LLM_ADAPTER_ALLOW_REMOTE_RUNTIME=true` is explicitly set
- document any model files or runtime caches created locally
- avoid sending uploaded image bytes to the model in the first version

## Observability

Adapter logs should be structured JSON and should include:

```text
request_id
status
adapter_version
model_runtime
model_name
duration_ms
normalized_field_count
warning_count
error_code
```

Logs should not include raw note text by default.

The backend should record adapter metadata in `inference` so users can see how a field was produced:

```text
engine_type: hybrid_rules_local_llm
engine_version: rules-0.1+llm-adapter-0.1
field.source: rules | observed_text | local_llm | fallback
field.confidence
field.explanation
```

## Health endpoints

Recommended adapter endpoints:

```text
GET /health/live
GET /health/ready
GET /v1/models/current
POST /v1/enrich/friction-event
```

Readiness should fail when the configured local runtime is unreachable or the configured model is unavailable. Liveness should fail only when the adapter process itself is unhealthy.

## Local setup guide for Ollama and Qwen

The implementation PR for this adapter must include installation and project setup documentation for Ubuntu and Windows 11. The setup guide should be tested manually on at least one platform before marking the adapter ready for user testing.

### Ubuntu setup

Recommended Ubuntu path:

```bash
curl -fsSL https://ollama.com/install.sh | sh
sudo systemctl enable ollama
sudo systemctl start ollama
ollama --version
ollama pull qwen3.6
ollama run qwen3.6
```

If Qwen3.6 is too large for the developer machine, the guide should document the smaller fallback model:

```bash
ollama pull qwen3:8b
ollama run qwen3:8b
```

If the developer has enough RAM/VRAM and wants to pin a specific Qwen3.6 size, they may use a concrete tag:

```bash
ollama pull qwen3.6:27b
# or, for stronger hardware:
ollama pull qwen3.6:35b
```

The guide should include optional GPU verification commands:

```bash
nvidia-smi
journalctl -e -u ollama
```

### Windows 11 setup

Recommended Windows 11 path:

1. Download and run `OllamaSetup.exe` from the official Ollama download page.
2. Open a new PowerShell window after installation so the updated PATH is available.
3. Verify Ollama is running:

```powershell
ollama --version
Invoke-WebRequest http://localhost:11434/api/tags
```

4. Pull and run the first supported Qwen model:

```powershell
ollama pull qwen3.6
ollama run qwen3.6
```

If Qwen3.6 is too large for the user's hardware, they may use `qwen3:8b` as the documented fallback. If they want a pinned Qwen3.6 size, they may use `qwen3.6:27b` or `qwen3.6:35b`.

### Qwen 3.6 model note

The default adapter setup should use the official Ollama library tag `qwen3.6`. This currently targets the Qwen3.6 family and is appropriate for the user's stronger local hardware. Because the model is large, the implementation documentation must also include `qwen3:8b` as a lower-resource fallback and keep the model tag fully configurable through `LOGARIFT_LLM_MODEL`.

### Project setup with the adapter enabled

After Ollama and the model are available locally, the project setup should be:

```bash
export LOGARIFT_LLM_ADAPTER_ENABLED=true
export LOGARIFT_LLM_ADAPTER_URL=http://localhost:8091
export LOGARIFT_LLM_RUNTIME_URL=http://localhost:11434
export LOGARIFT_LLM_MODEL=qwen3.6
docker compose up --build
```

The adapter implementation should also document the equivalent Windows PowerShell variables:

```powershell
$env:LOGARIFT_LLM_ADAPTER_ENABLED = "true"
$env:LOGARIFT_LLM_ADAPTER_URL = "http://localhost:8091"
$env:LOGARIFT_LLM_RUNTIME_URL = "http://localhost:11434"
$env:LOGARIFT_LLM_MODEL = "qwen3.6"
```

The project setup guide must include a smoke request to the local runtime before starting Logarift:

```bash
curl http://localhost:11434/api/chat \
  -d '{"model":"qwen3.6","messages":[{"role":"user","content":"Return only JSON: {\"ok\":true}"}],"stream":false,"format":"json"}'
```

If using the fallback model, replace `qwen3.6` with `qwen3:8b` in the smoke request.

### Official setup references

The implementation guide should link to these upstream pages so users can verify current installer and model details:

- Ollama Linux documentation: `https://docs.ollama.com/linux`
- Ollama Windows documentation: `https://docs.ollama.com/windows`
- Ollama API documentation: `https://docs.ollama.com/api`
- Ollama Qwen3.6 library page: `https://ollama.com/library/qwen3.6`
- Qwen3.6 upstream repository: `https://github.com/QwenLM/Qwen3.6`

## Testing strategy

Implementation should include:

- schema validation unit tests
- prompt construction golden tests
- backend merge-policy unit tests
- fake runtime integration tests with fixed JSON responses
- timeout and invalid JSON fallback tests
- Docker Compose smoke test with adapter disabled
- optional manual smoke test with a real local runtime

Tests must not require downloading a model by default.

## Rollout plan

Recommended phases:

1. Add backend client interface, configuration, and disabled-by-default feature flag.
2. Add adapter service skeleton with health endpoints and fake runtime mode for tests.
3. Add strict request/response schemas and validation.
4. Add backend merge policy and fallback tests.
5. Add local runtime integration behind explicit opt-in configuration.
6. Add UI metadata display for engine source and confidence.
7. Add Ubuntu and Windows 11 setup documentation for Ollama plus Qwen model selection and project environment variables.

