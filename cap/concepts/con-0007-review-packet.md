# Review packet

## Metadata

- context: ctx-0002

## Definition

The material an agent reads to assess one Entity (con-0001) against the conventions:
the entity's own content, and a checklist of questions particular to its kind. For a
capability the packet also embeds its Capability bundle (con-0006), so the assessor sees the
invariants and design alongside the capability. A review packet lets an agent judge an
entity and find documentation gaps without the tool making any judgement itself.

## States

A review packet has no lifecycle state. It is assembled on demand for one entity.

## Relationships

A review packet is assembled from the Model (con-0004) by reviewing an entity
(cap-0008). For a capability it embeds that capability's Capability bundle (con-0006).
