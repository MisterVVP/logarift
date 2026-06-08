# Logarift Ollama Modelfiles

These Modelfiles create optional Ollama model aliases tuned for Logarift's local LLM adapter response contract. They are intentionally aligned with `POST /v1/enrich/friction-event`: the model should return only `fields` and `warnings`, while the Go adapter adds adapter metadata and validates every field before the backend merges anything into canonical event data.

The Modelfiles adapt the external suggestion to Logarift's domain constraints. In particular, they do **not** include `suggested_next_action`, coaching, or recommendation fields because the initial release local adapter must enrich friction events, not advise users or rate productivity.

## Fast lower-resource model

```bash
ollama pull qwen3:8b
ollama create logarift-enricher-qwen3-8b -f llm-adapter/modelfiles/logarift-enricher-qwen3-8b.Modelfile
export LOGARIFT_LLM_MODEL=logarift-enricher-qwen3-8b
```

## Optional stronger 14B model

Use this after checking adapter warnings and response-shape diagnostics if the 8B alias remains too conservative on simple notes.

```bash
ollama pull qwen3:14b
ollama create logarift-enricher-qwen3-14b -f llm-adapter/modelfiles/logarift-enricher-qwen3-14b.Modelfile
export LOGARIFT_LLM_MODEL=logarift-enricher-qwen3-14b
```

## Stronger Qwen3.6 model

```bash
ollama pull qwen3.6
ollama create logarift-enricher-qwen36 -f llm-adapter/modelfiles/logarift-enricher-qwen36.Modelfile
export LOGARIFT_LLM_MODEL=logarift-enricher-qwen3-14b
# or:
export LOGARIFT_LLM_MODEL=logarift-enricher-qwen36
```

## Direct Ollama smoke test

Use this to inspect the model alias before routing through Logarift. Replace the model name with `logarift-enricher-qwen3-14b` or `logarift-enricher-qwen36` if using a stronger alias.

```bash
curl http://localhost:11434/api/chat \
  -d '{
    "model": "logarift-enricher-qwen3-8b",
    "stream": false,
    "format": "json",
    "messages": [
      {
        "role": "user",
        "content": "{\"allowed_values\":{\"workflow_stage\":[\"planning\",\"local_development\",\"build\",\"test\",\"code_review\",\"merge\",\"deployment\",\"operation\",\"debugging\",\"documentation\",\"coordination\",\"learning\"],\"friction_layer\":[\"technical\",\"temporal\",\"cognitive\",\"social_process\",\"motivational\",\"environmental\"],\"friction_type\":[\"slow_feedback\",\"failed_feedback\",\"unclear_error\",\"missing_documentation\",\"ambiguous_ownership\",\"interruption\",\"waiting_for_review\",\"waiting_for_ci\",\"context_switch\",\"rework\",\"tooling_failure\",\"environment_setup\",\"coordination_overhead\",\"decision_blocked\"]},\"deterministic_baseline\":{\"workflow_stage\":\"debugging\",\"friction_layer\":\"technical\",\"friction_type\":\"unclear_error\"},\"observed\":{\"plain_text\":\"Spent too much time debugging why the C++ math engine only prints startup logs. No visibility into calculations, hard to verify formulas.\"}}"
      }
    ]
  }'
```

## Adapter configuration

After creating an alias, point the adapter to it:

```bash
export LOGARIFT_LLM_ADAPTER_ENABLED=true
export LOGARIFT_LLM_MODEL=logarift-enricher-qwen3-8b
# or:
export LOGARIFT_LLM_MODEL=logarift-enricher-qwen3-14b
# or:
export LOGARIFT_LLM_MODEL=logarift-enricher-qwen36
```

The backend still validates ontology values, confidence gates, numeric ranges, and unknown fields. Deterministic rules remain the fallback when the adapter is disabled, slow, unavailable, or returns unusable output.
