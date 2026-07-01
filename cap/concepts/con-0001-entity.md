# Entity

## Metadata

- context: ctx-0001

## Definition

A documented thing of one kind, written as a single Markdown file with a heading for
each section its kind needs. The kinds are bounded context, concept, capability,
invariant, specification, verification, scenario, ADR, and task. An entity carries an
Identifier (con-0002) and links to other entities through References (con-0003).

## States

A capability or task carries a status: draft, proposed, in-progress, or done. The
other kinds have no lifecycle state.

## Relationships

An entity is named by exactly one Identifier (con-0002). An entity links to other
entities through References (con-0003) written in its sections. The set of all loaded
entities forms the Model (con-0004).
