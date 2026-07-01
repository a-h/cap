# Reference

## Metadata

- context: ctx-0001

## Definition

A link from one Entity (con-0001) to another, written as a bullet under a section
heading. A reference is either an Identifier (con-0002), such as `cap-0003`, or a
Markdown link whose text or target carries one. A capability references its context,
invariants, specifications, ADRs, scenarios, verification, and tasks this way. A
reference that names an entity that does not exist is a dangling reference, which
loading (cap-0002) records as a problem.

## States

A reference has no lifecycle state.

## Relationships

A reference is written in the section of one Entity (con-0001) and carries an
Identifier (con-0002) that names another. A many-to-many reference, such as the link
between an invariant and a capability, is declared on both entities (inv-0003).
