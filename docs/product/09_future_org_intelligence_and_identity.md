# Future Organisation Intelligence and Identity

## Purpose

This document captures future features that are intentionally not part of the initial public release: LLM/ML-assisted friction location, organisation and team inference, and optional SSO access gates.

These features must extend Logarift's non-surveillance posture instead of weakening it.

## Future Feature: LLM/ML-Assisted Friction Location

A future enrichment capability may use LLMs, machine-learning classifiers, or a hybrid rules-plus-model pipeline to estimate where recurring friction is likely located.

Possible outputs include:

- affected workflow stage
- likely platform, service, repository, or tool family
- probable owning team or support group
- impacted organisation area
- repeated dependency boundary
- documentation or onboarding gap

The feature should produce hypotheses for investigation, not evidence about individual performance.

## Required Guardrails

Future organisation intelligence must follow these constraints:

- keep raw logs anonymous by default
- avoid creating a stable person identifier
- avoid predicting who reported an event
- avoid predicting who caused an event
- use minimum cohort sizes before showing team or organisation aggregates
- prefer system, workflow, and ownership-boundary signals over person-level signals
- show confidence and uncertainty clearly
- preserve deterministic fallbacks when model output is low-confidence
- record model version, prompt version, and merge decisions for auditability
- never use private notes for manager-facing raw views by default

## Product UX Direction

The UI should frame model output as triage assistance:

```text
Likely area to investigate: build pipeline / frontend platform
Confidence: medium
Reason: repeated mentions of package installation delay, CI cache misses, and local environment drift
Suggested next step: inspect build tooling and onboarding docs with the owning platform team
```

The UI should not frame model output as attribution:

```text
Bad: Alice is slow because frontend builds are slow.
Bad: Team X caused developer Y's delay.
Bad: Rank teams by productivity.
```

## Future Feature: Optional SSO Access Gate

A later release may add an access gate for organizations that require private service entry through existing identity infrastructure.

Potential providers and protocols:

- Microsoft Entra ID
- AWS IAM Identity Center
- Google Cloud Identity
- generic OIDC
- generic SAML through a reverse proxy or gateway

SSO should answer only whether someone may access the deployment. It should not require Logarift to attach identity to friction events, dashboards, scores, or exports.

## Identity Boundary

If SSO is added, the preferred boundary is:

```text
identity provider or gateway authenticates request
Logarift receives only the minimum access signal needed to serve the app
Logarift stores friction events without user identity
aggregate analytics remain anonymous
```

Avoid adding application roles unless a concrete administrative operation requires them. If administrative roles are needed later, keep them separate from event authorship and analytics.

## Out of Scope for This Future Area

These capabilities should remain out of scope:

- individual developer profiles
- per-person productivity scores
- manager views into private timelines
- identity-linked friction history
- automated performance-evaluation evidence
- hidden ingestion from IDEs, chat, calendars, browsers, or repositories

## Open Product Questions

Before implementing this future area, decide:

- minimum cohort size for team and organisation aggregates
- whether organisation-area labels are configured manually, inferred, or both
- how people can correct wrong model suggestions without revealing identity
- what export controls are required for anonymous-but-sensitive notes
- whether SSO is handled inside Logarift or delegated to an ingress, gateway, or service mesh
