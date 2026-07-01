# List entities of a kind

## Metadata

- context: ctx-0002
- status: done

## Description

List every Entity (con-0001) of one kind in the Model (con-0004), each shown by its
Identifier (con-0002) and title, ordered by identifier.
This is how an author finds the identifier of an entity to reference, or reviews what
of a kind already exists before scaffolding another.

## Scope

In scope:

- Listing the entities of one kind by identifier and title.
- Marking an inline entity as inline in the listing.
- Rendering the list as text or JSON.

Out of scope:

- Showing an entity's links, which showing an entity (cap-0005) does.

## Actors

- Authors, who list a kind to find an identifier or survey what exists.

## Invariants

- Every entity of the requested kind appears in the listing exactly once, ordered by identifier.

## Specifications

### Listing

- A listing entry carries the identifier, the title, and whether the entity is inline.
- Inline entities, owned by a capability rather than held in their own file, are marked
  as inline so they are not mistaken for shareable entities.

## ADRs

## Scenarios

## Verification

- query/query_test.go

## Tasks
