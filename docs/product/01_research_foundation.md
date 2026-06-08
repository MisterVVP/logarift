# Research Foundation

## Purpose

This document summarizes the public research ideas that inform Logarift.

The project may use public scientific and industry research as conceptual background, but the product must maintain its own original terminology, ontology, schemas, formulas, user experience, and implementation.

## Developer Experience Friction

Developer Experience friction refers to obstacles that increase effort, delay feedback, interrupt concentration, reduce confidence, or create rework during software development.

Examples include:

- slow builds
- flaky tests
- unclear error messages
- missing documentation
- waiting for code review
- ambiguous ownership
- interrupted focus
- broken local environments
- repeated context switching
- unclear requirements

The important product assumption is that friction is not only about elapsed time. It also affects attention, cognitive load, motivation, recovery time, and future error probability.

## Cognitive Load Theory

Cognitive Load Theory distinguishes between different demands placed on working memory.

For this project, the most relevant idea is that unnecessary complexity consumes limited mental capacity.

Developer friction can increase extraneous cognitive load when a developer has to spend attention on accidental complexity rather than the intended engineering task.

Examples:

- deciphering unclear logs
- fighting local setup
- understanding undocumented behavior
- switching tools repeatedly
- recovering context after interruption

Product implication:

Logarift should estimate cognitive load pressure, not only time lost.

## Interruption and Resumption

Software development often requires maintaining complex mental state. Interruptions can cause developers to lose context and spend additional time resuming work.

Relevant concepts:

- interruption cost
- resumption lag
- context reconstruction
- attention residue
- task switching overhead

Product implication:

The initial release should track interruption count and resume time as first-class friction fields.

## Flow Theory

Flow describes a state of deep engagement where challenge and skill are well matched.

Developer flow can be disrupted by:

- unnecessary waiting
- frequent interruptions
- ambiguous next steps
- unstable tools
- noisy communication
- excessive context switching

Product implication:

The system should later model flow stability, but initial release should start by tracking interruption and recovery signals.

## Self-Determination Theory

Self-Determination Theory emphasizes autonomy, competence, and relatedness as important psychological needs.

Developer friction can harm these needs:

- autonomy: blocked by unclear ownership or excessive approval gates
- competence: harmed by confusing errors or unstable tools
- relatedness: harmed by poor coordination or lack of support

Product implication:

Friction should include motivational and social-process dimensions, not only technical categories.

## Prospect Theory and Loss Asymmetry

Behavioral economics suggests that losses can feel more salient than equivalent gains.

In a Developer Experience context, a negative friction event may have a stronger effect on perceived productivity and motivation than a similarly sized improvement.

Product implication:

Future scoring models may weight negative events asymmetrically. initial release should capture emotion valence to support later calibration.

## Queueing Theory

Many engineering workflows behave like queues:

- CI jobs waiting for runners
- pull requests waiting for review
- deployment approvals
- support requests
- platform team intake

When utilization approaches capacity, wait time can grow nonlinearly.

Product implication:

The system should eventually model systemic drag using queueing concepts. initial release starts with a simpler systemic drag estimate based on wait time versus active work time.

## Little's Law

Little's Law relates work in progress, throughput, and cycle time.

In simplified form:

```text
WIP = throughput * cycle_time
```

Product implication:

Future versions can use this principle to reason about delivery system congestion.

## Survival Analysis

Survival analysis models time until an event occurs.

In this project, possible applications include:

- time until friction is resolved
- time until task abandonment
- time until repeated friction recurs
- time until intervention effect appears

Product implication:

Survival analysis is out of initial release scope but belongs to future advanced models.

## Bayesian Updating

Bayesian inference allows beliefs to be updated as new evidence appears.

Possible future uses:

- estimate probability that a friction source is systemic
- update confidence in intervention effectiveness
- personalize model parameters
- separate noise from recurring patterns

Product implication:

Bayesian modeling is out of initial release scope. initial release formulas should be deterministic and explainable.

## Markov and State-Transition Modeling

Developer work can be approximated as transitions between states such as:

- focused
- interrupted
- waiting
- blocked
- recovering
- coordinating

Product implication:

A future model may estimate flow stability from transition probabilities. initial release should only collect enough data to support later state modeling.

## Software Engineering Productivity Frameworks

Public industry frameworks such as DORA and SPACE emphasize that software productivity should be understood through multiple dimensions rather than a single metric.

Relevant dimensions include:

- delivery performance
- satisfaction
- activity
- communication
- efficiency
- flow

Product implication:

Logarift should avoid one-dimensional productivity scoring. It should report multiple friction dimensions and explain their meaning.

## Research-to-Product Boundary

The project may use public research as inspiration.

The project must not:

- copy proprietary taxonomies
- copy internal company dashboards
- copy private survey instruments
- represent initial release formulas as scientifically validated universal laws
- claim causal certainty from descriptive data

The project should:

- use original terminology
- keep formulas explainable
- document assumptions
- allow parameter changes later
- separate observed data from inferred scores
