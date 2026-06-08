# Product Vision

## Problem Statement

Developer Experience friction is often visible only after it has already caused delay, frustration, quality issues, or delivery drag.

Common examples include slow feedback loops, unclear ownership, unstable development environments, confusing errors, repeated context switching, missing documentation, waiting for reviews, and rework caused by ambiguous requirements.

Most teams already experience these problems, but they usually measure them indirectly through delivery metrics, surveys, retrospectives, or anecdotal complaints. These approaches are useful, but they often miss the moment-by-moment structure of friction and how small frictions compound over time.

Logarift is an anonymous, centrally deployable system for recording, analyzing, and explaining Developer Experience friction across a tech organization. It can still run locally for DevEx platform developers, contributors, and demos, but the product concept is an organization-available service rather than a personal tracker.

The central idea is:

```text
Friction is not only a logged event.
Friction is a compounding signal that affects cognitive load, flow stability, and systemic delivery drag.
```

## Target Users

Primary users:

- software developers across a tech organization
- platform engineers
- Developer Experience engineers
- technical leads
- engineering managers
- internal tools teams
- DevEx platform developers operating the service

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
- anonymous shared analytics
- mathematical scoring
- time-dependent friction modeling
- explainable dashboards

## Differentiation

Logarift should not be a generic productivity tracker.

The differentiating ideas are:

1. **Anonymous organization trust model**

   The service should be available to the whole tech organization without requiring Logarift-managed user accounts. Friction logs are anonymous by default, and deployment owners should minimize access barriers while keeping infrastructure private.

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

## Deployment Principle

The default product posture is centralized private deployment for a tech organization:

- containers for each runtime component
- Kubernetes/Helm for shared installations
- Go backend API
- React + Vite frontend
- MongoDB
- C++ math engine
- Valkey-backed asynchronous enrichment when enabled
- optional LLM adapter

Local Docker Compose remains a supported path for DevEx platform developers, contributors, demos, and safe offline testing. No Logarift cloud account is required.

## Non-Surveillance Principle

Logarift is designed for anonymous friction reporting, self-reflection, and workflow improvement.

It must not become:

- an employee monitoring system
- an individual productivity ranking tool
- hidden telemetry
- a manager dashboard for inspecting private developer timelines
- a tool for identifying who caused or experienced friction

Future team and organisation features should use aggregation, anonymity-preserving design, minimum cohort sizes, and clear labeling that results are decision-support signals rather than performance evidence.

## Initial Public Release Vision

The initial public release should allow anonymous users of a centrally deployed instance, and local DevEx platform developers running Docker Compose, to:

- create friction events
- classify events using the original ontology
- link events to work sessions and goals
- compute basic friction scores
- view an anonymous shared dashboard
- export data as JSON

## Long-Term Vision

Future versions may support:

- local Git and CI importers
- intervention simulation
- team-level aggregate analysis with minimum cohort protections
- LLM/ML-assisted friction location that suggests likely affected systems, teams, or organisation areas from anonymous logs
- optional SSO gate for deployment access using Entra ID, AWS IAM Identity Center, Google Cloud Identity, or generic OIDC/SAML without introducing per-person productivity views
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
