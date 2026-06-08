# Product Vision

## Problem Statement

Developer Experience friction is often visible only after it has already caused delay, frustration, quality issues, or delivery drag.

Common examples include slow feedback loops, unclear ownership, unstable development environments, confusing errors, repeated context switching, missing documentation, waiting for reviews, and rework caused by ambiguous requirements.

Most teams already experience these problems, but they usually measure them indirectly through delivery metrics, surveys, retrospectives, or anecdotal complaints. These approaches are useful, but they often miss the moment-by-moment structure of friction and how small frictions compound over time.

Logarift is a local-first system for recording, analyzing, and explaining Developer Experience friction.

The central idea is:

```text
Friction is not only a logged event.
Friction is a compounding signal that affects cognitive load, flow stability, and systemic delivery drag.
```

## Target Users

Primary users:

- individual software developers
- platform engineers
- Developer Experience engineers
- technical leads
- engineering managers
- internal tools teams
- small engineering teams improving their own workflows

Secondary users:

- DevOps/SRE teams
- productivity researchers
- open-source maintainers
- consultants analyzing engineering workflows

## Product Goal

The product should help users answer:

- Where does developer friction appear?
- Which friction types repeat most often?
- Which frictions cause the greatest time loss?
- Which frictions appear small individually but compound over time?
- Which workflow stages create the most cognitive or systemic drag?
- Which interventions are likely to reduce friction meaningfully?

## Core Product Idea

Logarift treats friction as a socio-technical signal.

A friction event may come from a tool, process, environment, person-to-person dependency, unclear knowledge boundary, or interruption. The product should preserve this context without becoming a surveillance or individual performance tool.

The system should combine:

- manual developer logging
- structured event metadata
- local analytics
- mathematical scoring
- time-dependent friction modeling
- explainable dashboards

## Differentiation

Logarift should not be a generic productivity tracker.

The differentiating ideas are:

1. **Local-first trust model**

   The user owns the data. initial release data stays local.

2. **Friction as a dynamic signal**

   The system models accumulation, recovery, clustering, and compounding instead of only counting events.

3. **Cognitive and systemic dimensions**

   The product considers both perceived cognitive load and measurable delivery drag.

4. **Explainable formulas**

   initial release scoring should be transparent and configurable later.

5. **Intervention-oriented analytics**

   The long-term goal is not only to report friction, but to estimate which improvements are worth doing.

6. **Non-surveillance by design**

   The system must not rank individual developers or silently collect behavioral data.

## Local-First Principle

The initial release runs locally with:

- Go backend API
- React + Vite frontend
- MongoDB
- C++ math engine
- Docker Compose

No cloud account is required.

## Non-Surveillance Principle

Logarift is designed for self-reflection and workflow improvement.

It must not become:

- an employee monitoring system
- an individual productivity ranking tool
- hidden telemetry
- a manager dashboard for inspecting private developer timelines

Future team features should use aggregation, anonymization or pseudonymization, minimum cohort sizes, and explicit opt-in.

## initial release Vision

The initial release should allow a single local user to:

- create friction events
- classify events using the original ontology
- link events to work sessions and goals
- compute basic friction scores
- view a local dashboard
- export data as JSON

## Long-Term Vision

Future versions may support:

- local Git and CI importers
- intervention simulation
- team-level aggregate analysis
- advanced mathematical models
- plugin-based data ingestion
- IDE integrations
- GitHub/GitLab/Jira/Linear integrations

The long-term product direction is:

```text
From friction logging to friction modeling.
From dashboards to intervention decisions.
From anecdotes to explainable socio-technical signals.
```
