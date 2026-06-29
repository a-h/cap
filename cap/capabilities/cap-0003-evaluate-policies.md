# Evaluate policies

## Metadata

- context: ctx-0001
- status: done

## Description

Evaluate policies to decide whether a given action is permitted. The outcome is a
permit or deny decision together with the reason for it. This capability is
independent of how policies are authored, stored, or surfaced in any interface.

## Scope

In scope:

- Combining a policy set and request inputs into a decision.
- Returning the reason a decision was reached.

Out of scope:

- Authoring or storing policy definitions.
- Enforcing a decision once made.

## Actors

- Services that must check whether an action is permitted before performing it.

## Invariants

- Policies are evaluated consistently.
- An empty policy set denies by default.
- An explicit deny overrides any permit.
- A decision is a pure function of the policy set and the request inputs.

## Specifications

### Policy evaluation design

- Policies compile to a decision tree, evaluated in a single pass over the request.
- Deny-overrides is enforced by ordering deny nodes ahead of permit nodes, so the
  first match for an explicit deny wins.
- The compiled tree is pure and cached per policy-set version, which is how
  consistent and side-effect-free evaluation is realised.

## Scenarios

- [scn-0001](../scenarios/scn-0001-claim-approval.md)

## Verification

- internal/policy/eval_test.go
- internal/policy/expiry_test.go
