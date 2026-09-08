# Generate code from a capability

## Description

An agent generates code and tests for a capability from its documented invariants and
design, then fills any documentation gap it finds by comparing the source against the
model. The path crosses two capabilities: providing context to the agent
([cap-0007](../capabilities/cap-0007-provide-context-to-an-agent.md)), which gives it
the rules to build from, and reviewing an entity
([cap-0008](../capabilities/cap-0008-review-an-entity.md)), which surfaces the gaps.

## Steps

- An author scaffolds and fills a capability, its invariants, and its specification.
- The agent bundles the capability context
  ([cap-0007](../capabilities/cap-0007-provide-context-to-an-agent.md)) to read the
  capability, its context, its invariants, and its design in one report.
- The agent generates the code and tests that uphold the invariants.
- The agent reviews the entity
  ([cap-0008](../capabilities/cap-0008-review-an-entity.md)) to compare the source
  against the model and reports gaps, such as an invariant with no verification.

## Capabilities

- [cap-0007 Provide context to an agent](../capabilities/cap-0007-provide-context-to-an-agent.md)
- [cap-0008 Review an entity](../capabilities/cap-0008-review-an-entity.md)
