# Name the design, for example "Policy evaluation design"

<!-- A specification is the design that realises a set of invariants: how the parts fit together, including implementation detail. It describes how, and does not restate the behaviour rules, which belong in the invariants. It specifies one or more capabilities, or a whole bounded context. -->

## Specifies (optional)

List the capabilities this specification is the design of, one identifier per bullet,
for example `cap-0003`. A specification may specify several capabilities. To specify a
whole bounded context instead, name the context here, for example `ctx-0001`. Declare a
capability link on both entities: name the capabilities here, and name this
specification under each capability's Specifications section. Either file then shows the
relationship on its own, and cap validate warns when only one side names the other.

- cap-0003

## Description

Describe the overall design that realises its invariants: how the pieces fit together,
including implementation detail. Do not restate the behaviour rules; those are the
invariants.

## Design

- How the parts combine to uphold the invariants.
