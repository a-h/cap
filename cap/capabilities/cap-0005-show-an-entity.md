# Show an entity

## Metadata

- context: ctx-0002
- status: done

## Description

Show one Entity (con-0001): its title, the entities it links to, and the entities
that link to it. Showing an entity answers "what does this
connect to" in both directions from a single Identifier (con-0002), so an author can
follow the Model (con-0004) from any starting point.

## Scope

In scope:

- Resolving the children an entity links to and the parents that link to it.
- Rendering the entity, its children, and its parents as text or JSON.

Out of scope:

- Expanding the links recursively, which graphing the model (cap-0006) does.
- Assembling everything an agent needs from a capability, which bundling a capability
  context (cap-0007) does.

## Actors

- Authors, who show an entity to see its immediate links in both directions.

## Invariants

- Both the entities an entity links to and the entities that link to it are reported.

## Specifications

### Children and parents

- Children are the entities an entity links to downward, including the reverse links a
  file-backed invariant or specification declares back to a capability.
- Parents are the entities that link to it: its context, the scenarios that name it,
  the capabilities that reference it, and the specifications that specify it.

## ADRs

## Scenarios

## Verification

- query/query_test.go

## Tasks
