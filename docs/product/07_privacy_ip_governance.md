# Privacy and IP Governance

## Purpose

This document defines privacy and intellectual property principles for Logarift.

The system should be useful for Developer Experience improvement without becoming a surveillance tool or copying proprietary work.

## Privacy Principles

### Anonymous by Default

Initial release events should not carry Logarift-managed user identity.

The product should not require per-person accounts, event ownership, or authorization rules to deliver its core value.

Centralized private deployments are acceptable when infrastructure owners keep access controlled and the application remains anonymous by default. Local Docker Compose remains available for DevEx platform developers and contributors.

No Logarift cloud account is required.

No Logarift-controlled cloud sync is included in initial release.

### Explicit User Control

People and deployment operators should have explicit controls over:

- what is logged
- what is exported
- what is deleted
- whether any future integrations are enabled
- whether a centralized deployment is exposed only through approved private access paths

### No Hidden Telemetry

The system must not silently collect or identity-link:

- IDE activity
- keystrokes
- chat messages
- calendar events
- browser history
- private repository content

### No Individual Productivity Ranking

The product must not rank developers by productivity.

Future team and organisation features must avoid individual leaderboards, private timelines, identity-linked histories, and manager views that identify who experienced or caused friction.

### Private Notes Stay Local

Free-text notes may contain sensitive details.

Initial release notes are stored in MongoDB for the deployment.

Future export, team, and organisation-intelligence features should treat notes as sensitive by default and should not expose raw notes in manager-facing aggregate views.

### Data Deletion

Deletion should be available at the entity level in the initial release.

Future versions should include deployment-level retention controls, full local data reset for developer mode, and privacy-aware centralized reset/export workflows.

### Future Team, Organisation, and SSO Features

If aggregate team or organisation features are added later, they should use:

- aggregation
- anonymity-preserving design
- minimum cohort sizes
- no private timeline access for managers
- no raw notes in team dashboards by default
- no model output that identifies who reported or caused friction
- clear confidence and uncertainty labels for LLM/ML-inferred systems, teams, or organisation areas

If SSO is added later, it should be an access gate through providers such as Microsoft Entra ID, AWS IAM Identity Center, Google Cloud Identity, or generic OIDC/SAML. SSO should not require Logarift to attach identity to friction events, scores, or dashboards.

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
