# Capability bundle

## Metadata

- context: ctx-0002

## Definition

Everything relevant to one capability (con-0001), resolved from References (con-0003)
into full content and gathered into a single reading unit: the capability, its
context, and the full text of every invariant, specification, ADR, verification,
scenario, and task linked to it, including those that name the capability back. A
capability bundle is the unit an agent reads to generate code or tests, so it does not
have to walk the Model (con-0004) itself and cannot miss a linked invariant.
References that do not resolve are listed rather than dropped.

## States

A capability bundle has no lifecycle state. It is assembled on demand for one
capability.

## Relationships

A capability bundle is assembled from the Model (con-0004) by providing context to an
agent (cap-0007). Unlike a Graph (con-0005), a capability bundle roots only at a
capability, is one level deep, and holds resolved content rather than structure. A
Review packet (con-0007) for a capability embeds its bundle.
