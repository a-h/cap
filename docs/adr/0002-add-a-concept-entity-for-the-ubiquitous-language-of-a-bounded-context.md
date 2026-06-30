# 2. Add a concept entity for the ubiquitous language of a bounded context

Date: 2026-06-30

## Status

Accepted

## Context

A capability is named as a verb plus a noun, for example "Evaluate policies". The noun
refers to a thing in the domain, a policy, a decision, a request, but the model had no
entity for that thing. The vocabulary of a bounded context, its ubiquitous language,
was left implicit in capability names and prose.

Two problems followed. First, with no definition of the domain things, authors and
agents reach for undefined terms such as "world state" that have no shared meaning.
Second, when an agent is asked to populate the capabilities of a context, it tends to
produce a few broad capabilities such as "Manage policies" rather than the atomic set
the model wants, because nothing prompts it to enumerate the things in the domain and
the verbs that act on each.

## Decision

Add a `concept` entity, identifier prefix `con`, stored under `concepts/`. A concept
records a thing in a bounded context's language: its definition, optionally its
lifecycle states and its relationships to other concepts. A concept belongs to one
context, the only structural reference it carries, validated like a capability's
context.

Concepts appear beneath their context in `cap graph` and `cap show`, before the
context's capabilities, so the things a context operates on are read alongside the
abilities that act on them.

References to a concept from other entities are textual, not a link section: an entity
mentions a concept by name in its prose and tags the name with the concept's
identifier, for example "Policy (con-0001)". The review checklists carry the
corresponding judgement prompts: that each thing a context operates on is defined as a
concept, that a recurring undefined noun signals a missing concept, that a concept's
name is tagged where it is used, and that the capabilities acting on each concept are
atomic rather than bundled.

Alternatives considered:

- A glossary section inside the context document rather than a first-class entity.
  Rejected because a concept is then not addressable by identifier, cannot be linked
  or traced, and does not appear in `list`, `show`, or `graph` like the other kinds.
- Naming the entity "entity" or "domain entity". Rejected because "entity" is already
  the umbrella term for every kind in the model, so the name would be ambiguous, and
  "domain entity" carries domain-driven-design connotations (identity, value objects,
  aggregates) narrower than the glossary term intended here. "Concept" is the neutral
  word for a thing in the ubiquitous language.
- A `Concepts` link section on capabilities, so a capability names the concepts it
  acts on. Deferred: the inline, name-with-identifier convention keeps concepts a
  matter of language rather than another link graph to maintain, and the review step
  catches untagged and undefined uses without new schema.

## Consequences

The vocabulary of a bounded context is written down and addressable. Defining the
concepts of a context gives an agent a basis for decomposition: enumerate the things,
then the verbs on each, which yields atomic capabilities at the right altitude rather
than broad "manage" capabilities.

Concept references are not machine-checked. Because a concept is referenced by name in
prose rather than by a link section, `cap validate` cannot confirm that a tag is
present or correct; the review checklist prompts are the means of catching untagged
and missing concepts, performed by an assessor rather than the validator.

The existing commands extend to the new kind without new commands: `cap new concept`,
`cap list concepts`, `cap show con-0001`, and `cap review con-0001` all work through
the kind-generic machinery.
