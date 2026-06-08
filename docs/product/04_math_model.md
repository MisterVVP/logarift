# Math Model

## Purpose

This document defines the first deterministic mathematical model for Logarift.

The initial release model is intentionally simple, explainable, and implementation-friendly.

These formulas are not universal scientific truths. They are product hypotheses inspired by public research on cognitive load, interruptions, flow, queueing, and socio-technical systems.

Parameters should become configurable in later versions.

## Design Principles

The initial release math model should be:

- deterministic
- explainable
- easy to test
- stable for small datasets
- based on event fields already collected in initial release
- separated from raw observed data
- versioned

## Input Fields

The initial release formulas use:

```text
severity_self
cognitive_load_self
time_lost_minutes
resume_time_minutes
interruption_count
recovery_minutes
timestamp_start
timestamp_end
workflow_stage
friction_layer
friction_type
```

If `recovery_minutes` is not available in the initial release UI, it should default to `0`.

## Score 1: Cognitive Load Accumulator

The Cognitive Load Accumulator estimates accumulated mental pressure over time.

Formula:

```text
CLA_t = decay * CLA_(t-1)
        + severity_weight
        + cognitive_load_weight
        + interruption_weight
        + resume_penalty
        - recovery_bonus
```

Suggested initial release parameters:

```text
decay = 0.85
severity_weight = severity_self * 1.2
cognitive_load_weight = cognitive_load_self * 1.5
interruption_weight = interruption_count * 2.0
resume_penalty = log(1 + resume_time_minutes)
recovery_bonus = recovery_minutes * 0.3
```

Implementation notes:

- process events in timestamp order
- clamp final score to minimum `0`
- store model version with output
- use natural logarithm

Example output field:

```json
{
  "cla": 37.4
}
```

## Score 2: Friction Compounding Index

The Friction Compounding Index estimates whether friction events cluster close together.

Formula:

```text
FCI = sum(event_weight * exp(-delta_minutes / half_life))
```

Suggested initial release parameters:

```text
half_life = 90 minutes
event_weight = severity_self + cognitive_load_self + log(1 + resume_time_minutes)
```

Where:

```text
delta_minutes = minutes between event timestamp and scoring period end
```

Interpretation:

- recent events count more
- high severity events count more
- high cognitive load events count more
- long resumption events count more
- clustered friction produces a higher current score

Example output field:

```json
{
  "fci": 21.8
}
```

## Score 3: Systemic Drag Coefficient

The initial release Systemic Drag Coefficient estimates waiting burden relative to active work.

initial release formula:

```text
SDC = total_wait_time_minutes / max(total_active_work_minutes, 1)
```

Suggested classification of wait-like types:

```text
waiting_for_review
waiting_for_ci
decision_blocked
coordination_overhead
```

Interpretation:

- `0.0` means no recorded waiting drag
- `0.5` means wait time equals half of active work time
- `1.0` means wait time equals active work time
- values above `1.0` indicate heavy systemic waiting burden

Future queueing formula:

```text
rho = arrival_rate / service_rate
SDC = rho / (1 - rho)
```

The queueing formula is not part of initial release implementation unless enough queue data exists.

Example output field:

```json
{
  "sdc": 0.42
}
```

## Score 4: Friction Cost Score

The Friction Cost Score estimates the local impact of one event.

Formula:

```text
FCS = severity_self
      * (1 + log(1 + time_lost_minutes))
      * (1 + 0.2 * interruption_count)
```

Interpretation:

- higher severity increases score
- time loss increases score nonlinearly
- interruptions amplify impact

Example event-level output:

```json
{
  "event_id": "evt_123",
  "fcs": 14.6
}
```

## Score Output

The math engine should return:

```json
{
  "model_version": "model-0.1",
  "period_start": "2026-06-01T00:00:00Z",
  "period_end": "2026-06-07T23:59:59Z",
  "scores": {
    "cla": 37.4,
    "fci": 21.8,
    "sdc": 0.42
  },
  "event_scores": [
    {
      "event_id": "evt_123",
      "fcs": 14.6
    }
  ],
  "top_contributors": [
    {
      "event_id": "evt_123",
      "reason": "high severity and long resume time"
    }
  ]
}
```

## Normalization

initial release does not require advanced normalization.

Future versions may normalize by:

- personal baseline
- workflow stage
- day of week
- session type
- team baseline

## Model Versioning

Every score snapshot must store:

```text
model_version
model_config_id
created_at
period_start
period_end
```

## Explainability

Each score should be explainable.

The scoring output should include top contributors where possible.

Example reasons:

```text
high severity
high cognitive load
long resume time
many interruptions
recent clustered event
wait-like friction type
```

## Constraints

Do not use black-box machine learning in initial release.

Do not claim causal certainty.

Do not rank developers.

Do not treat raw time loss as the only important signal.
