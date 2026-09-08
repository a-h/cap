# An entity has a canonical identifier

## Metadata

- capabilities:

## Description

Every Entity (con-0001) is named by one Identifier (con-0002) in canonical form: a
lowercase kind prefix and a zero-padded number. Scaffolding a new entity allocates the
next free number for its kind, so no two entities of a kind share an identifier, and
loading reads the identifier from the file name.

## Capabilities

- [cap-0001 Scaffold an entity](../capabilities/cap-0001-scaffold-an-entity.md)
- [cap-0002 Load the model](../capabilities/cap-0002-load-the-model.md)

## Rationale

The identifier is how every Reference (con-0003) points at an entity and how a report
addresses it. Two entities sharing an identifier, or an identifier that does not
follow the canonical form, would make references ambiguous and break lookup.
