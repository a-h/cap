# Provide context to an agent

## Metadata

- context: ctx-0002
- status: done

## Description

Give an agent the context it needs to develop or understand one capability (con-0001),
so it can write code or tests, or explain existing behaviour, without reading the
whole Model (con-0004). The capability gathers everything that constrains one
capability into a Capability bundle (con-0006): the capability itself, its context,
and the full text of every invariant, specification, ADR, verification, scenario, and
task linked to it, including those that name the capability back. An agent given this
bundle works from the rules the capability must uphold rather than guessing them, and
cannot miss a linked invariant. References (con-0003) that do not resolve are listed
rather than dropped.

## Scope

In scope:

- Resolving the capability's context and its linked invariants, specifications, ADRs,
  verification, scenarios, and tasks into their full form.
- Including the invariants and scenarios that name the capability back, not only those
  it names.
- Recording references that did not resolve.

Out of scope:

- Generating the code or tests, or explaining the feature, which the agent does from
  the bundle.
- The review checklist, which reviewing an entity (cap-0008) adds.

## Actors

- An agent developing or understanding a feature, which needs the capability's rules
  and design before it writes or explains code.

## Invariants

- A reference that does not resolve is reported in the bundle rather than dropped.

## Specifications

### Bundle assembly

- Invariants are collected from both the capability's list and the file-backed
  invariants that name it back, deduplicated and ordered by identifier.
- Scenarios are collected from both the capability's list and the scenarios that name
  it, so the bundle reflects links declared on either side.
- The context, specifications, ADRs, verification, and tasks are resolved from the
  capability's own references.

## ADRs

## Scenarios

- scn-0001

## Verification

- context/context_test.go

## Tasks
