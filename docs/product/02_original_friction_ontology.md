# Original Friction Ontology

## Purpose

This document defines the initial Logarift friction ontology.

The ontology is original project terminology. It should evolve through dogfooding and real usage.

The ontology must help describe developer friction without copying proprietary company taxonomies or survey instruments.

## Core Concepts

### Goal

A goal is a meaningful work outcome the developer is trying to achieve.

Examples:

- implement a feature
- fix a bug
- review a pull request
- investigate an incident
- write documentation
- deploy a service

### Session

A session is a bounded period of work related to one or more goals.

A session may contain multiple friction events.

Examples:

- morning feature work block
- debugging session
- release preparation session
- code review session

### Friction Event

A friction event is a specific moment or interval where progress becomes harder, slower, more confusing, or more cognitively expensive.

A friction event should be small enough to log quickly.

Examples:

- CI failed with unclear error
- local environment broke
- waited for review
- switched context due to interruption
- spent time searching for missing documentation

### Friction Episode

A friction episode is a cluster of related friction events that together represent a larger problem.

Example:

- failed build
- searched logs
- found outdated documentation
- asked for help
- waited for response

These may be separate events but one episode.

Episodes may be implemented after initial release.

### Workflow Stage

Workflow stage describes where the friction happened in the development workflow.

initial release workflow stages:

```text
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
```

### Friction Layer

Friction layer describes the broad nature of the friction.

initial release friction layers:

```text
technical
temporal
cognitive
social_process
motivational
environmental
```

Layer meanings:

- technical: tools, systems, code, infrastructure
- temporal: waiting, delay, slow feedback
- cognitive: confusion, excessive complexity, context loss
- social_process: ownership, review, coordination, decision-making
- motivational: frustration, confidence loss, reduced sense of progress
- environmental: workspace, interruptions, device/network conditions

### Friction Type

Friction type is a more specific classification.

initial release friction types:

```text
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

### Outcome

Outcome describes what happened after the friction event.

Suggested outcomes:

```text
resolved
worked_around
deferred
abandoned
escalated
still_open
```

### Recovery Signal

Recovery signal describes how costly it was to return to useful work.

initial release recovery fields:

```text
resume_time_minutes
interruption_count
recovery_minutes
```

## initial release Event Description

A friction event should capture:

```text
what happened
where it happened
why it mattered
how much time it cost
how much cognitive load it created
how difficult it was to resume work
what goal/session it affected
```

## Severity

Severity is the user's self-assessed impact of the friction event.

Suggested initial release scale:

```text
1 = barely noticeable
2 = minor annoyance
3 = meaningful slowdown
4 = major disruption
5 = severe blocker
```

## Cognitive Load

Cognitive load is the user's self-assessed mental effort caused by the event.

Suggested initial release scale:

```text
1 = low mental effort
2 = mild effort
3 = moderate effort
4 = high effort
5 = exhausting or deeply confusing
```

## Emotion Valence

Emotion valence captures affective direction.

Suggested initial release scale:

```text
-2 = strongly negative
-1 = negative
 0 = neutral
 1 = positive
 2 = strongly positive
```

Positive values may be useful when logging recovery, learning, or successful intervention events.

## Source

Source identifies where the event came from.

initial release sources:

```text
manual
seed
import
```

initial release uses `manual` and `seed`. Importers are future scope.

## Ontology Evolution

The ontology should evolve based on usage.

Rules for evolution:

- keep backward compatibility through schema_version
- avoid company-specific proprietary categories
- prefer plain language
- avoid blame-oriented labels
- preserve the distinction between observed fields and inferred scores
- keep classification fast enough for daily use
