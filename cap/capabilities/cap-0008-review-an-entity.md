# Review an entity

## Metadata

- context: ctx-0002
- status: done

## Description

Give an agent what it needs to assess one Entity (con-0001) against the conventions,
so it can find documentation gaps and weak entities the tool itself does not judge.
The capability assembles a Review packet (con-0007): the entity's own content and a
checklist of questions particular to its kind. For a capability the packet embeds its
Capability bundle (con-0006), so the assessor sees the invariants and design alongside
it. Given no identifier, a packet is assembled for every entity. Reviewing an entity is
how documentation gaps are found by comparing the source against the Model (con-0004).

## Scope

In scope:

- Reading an entity's content and building a checklist of review questions for its
  kind.
- Including the Capability bundle (con-0006) for a capability under review.
- Assembling a packet for one entity, or for every entity, rendered as text or JSON.

Out of scope:

- Making the assessment or editing the entity, which the agent and author do.
- Checking references mechanically, which validation (cap-0003) does.

## Actors

- An agent assessing an entity against the conventions and looking for gaps.

## Invariants

- A review packet carries a checklist of questions particular to the entity's kind.

## Specifications

### Packet assembly

- A packet carries the identifier, kind, title, source file, raw content, and the
  kind's checklist; a capability packet also carries its context bundle.
- Each kind's checklist encodes its naming and schema conventions, for example that a
  capability name is an imperative verb and noun and completes "the system can ___".

## ADRs

## Scenarios

- scn-0001
- scn-0002

## Verification

- review/review_test.go

## Tasks
