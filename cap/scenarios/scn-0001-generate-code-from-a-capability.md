# Generate code from a capability

## Description

An agent generates code and tests for a capability from its documented invariants and
design, then fills any documentation gap it finds by comparing the source against the
model. The path crosses two capabilities: providing context to the agent (cap-0007),
which gives it the rules to build from, and reviewing an entity (cap-0008), which
surfaces the gaps.

## Steps

- An author scaffolds and fills a capability, its invariants, and its specification.
- The agent bundles the capability context (cap-0007) to read the capability, its
  context, its invariants, and its design in one report.
- The agent generates the code and tests that uphold the invariants.
- The agent reviews the entity (cap-0008) to compare the source against the model and
  reports gaps, such as an invariant with no verification.

## Capabilities

- cap-0007
- cap-0008
