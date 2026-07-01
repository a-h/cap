# Identifier

## Metadata

- context: ctx-0001

## Definition

The name of an Entity (con-0001): a lowercase kind prefix and a zero-padded number,
such as `ctx-0001`, `cap-0003`, or `inv-0012`. The prefix determines the kind, so an
identifier is enough to know what an entity is and where its file lives. An inline
entity, defined within a capability rather than in its own file, has a synthesised
identifier of the form `cap-0003/inv-1`.

## States

An identifier has no lifecycle state.

## Relationships

An identifier names exactly one Entity (con-0001). A Reference (con-0003) carries an
identifier to point at the entity it names.
