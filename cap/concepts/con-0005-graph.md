# Graph

## Metadata

- context: ctx-0002

## Definition

The structural view of the Model (con-0004) produced by walking composition references
(con-0003) downward: a context and the capabilities beneath it, a capability and its
invariants, and so on. A graph carries an entity's identifier and title at each node,
whether each reference resolved, and whether a node repeats an ancestor, but not the
entities' content. It answers "what is the shape of this part of the model", for any
entity and to any depth.

## States

A graph has no lifecycle state. It is built on demand from the current model.

## Relationships

A graph is built from the Model (con-0004) by graphing the model (cap-0006) over the
References (con-0003) between Entities (con-0001). Unlike a capability bundle
(con-0006), a graph is recursive, roots at any entity, and shows structure rather than
resolved content.
