# Privacy and IP Governance

## Purpose

This document defines privacy and intellectual property principles for Logarift.

The system should be useful for Developer Experience improvement without becoming a surveillance tool or copying proprietary work.

## Privacy Principles

### Local-First by Default

initial release data stays local.

No cloud account is required.

No cloud sync is included in initial release.

### Explicit User Control

The user should control:

- what they log
- what they export
- what they delete
- whether any future integrations are enabled

### No Hidden Telemetry

The system must not silently collect:

- IDE activity
- keystrokes
- chat messages
- calendar events
- browser history
- private repository content

### No Individual Productivity Ranking

The product must not rank developers by productivity.

Future team features must avoid individual leaderboards.

### Private Notes Stay Local

Free-text notes may contain sensitive details.

initial release notes stay local in MongoDB.

Future export and team features should treat notes as sensitive by default.

### Data Deletion

The user should be able to delete local data.

initial release may implement deletion at the entity level.

Future versions should include full local data reset.

### Future Team Mode

If team mode is added later, it should use:

- explicit opt-in
- aggregation
- anonymization or pseudonymization
- minimum cohort sizes
- no private timeline access for managers
- no raw notes in team dashboards by default

## IP Principles

### Original Product Expression

The project owns its original:

- documentation
- ontology
- schemas
- formulas
- implementation
- UI design
- examples
- diagrams
- task definitions

### Public Research Boundary

The project may use public research as conceptual background.

Examples:

- cognitive load
- interruption recovery
- flow
- self-determination
- queueing theory
- Bayesian inference
- Markov models
- software productivity frameworks

Public scientific ideas are background knowledge, but the project should create its own implementation and expression.

### Do Not Copy Proprietary Work

Do not copy:

- proprietary taxonomies
- private dashboards
- company-internal metrics
- non-public survey instruments
- proprietary scoring systems
- exact wording from company materials

### Avoid Confusing Claims

Do not claim:

- that initial release formulas are scientifically validated universal metrics
- that descriptive scores prove causality
- that the system can measure individual developer productivity completely
- that friction scores should be used for performance evaluation

### Citation Practice

Research documents should cite public sources when specific claims are made.

Product implementation documents may describe original design decisions without copying source wording.

### Repo History as Design Record

Important design choices should be documented in repo history through:

- markdown docs
- task files
- architecture decision records
- commit messages

This helps establish original project development over time.

## Licensing Note

The repository license should be selected separately.

The license should match the intended open-source strategy and the author's goals around adoption, contribution, and commercial reuse.
