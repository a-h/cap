# Document existing code

## Description

An agent documents a system that already has source but no model, working from the
code toward the model rather than the other way round. It reads the source to find the
capabilities the system already has, scaffolds an entity for each, fills the entities
from what the code does, and uses validation and review to find where the model and
the source disagree. This is the path that produced cap's own self-documentation.

## Steps

- The agent reads the source to identify the bounded contexts, concepts, and
  capabilities the code already implements.
- The agent scaffolds an entity (cap-0001) for each context, concept, capability,
  invariant, and scenario it finds, and fills each from the source, citing the tests
  that verify it.
- The agent validates the model (cap-0003) to find dangling references, asymmetric
  links, and coverage gaps, and corrects them.
- The agent reviews an entity (cap-0008) against the conventions to find where a name
  reflects the mechanism rather than the use, or where the source asserts a behaviour
  the model does not yet record as an invariant.

## Capabilities

- cap-0001
- cap-0003
- cap-0008
