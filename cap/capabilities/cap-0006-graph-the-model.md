# Graph the model

## Metadata

- context: ctx-0002
- status: done

## Description

Produce a Graph (con-0005) of the composition beneath an Entity (con-0001): a
context and its capabilities, a capability and its invariants, and so on, down the
hierarchy. Given no identifier, graph the whole Model (con-0004) from its bounded
contexts. This is how an author sees the shape of a subtree, or the whole model, at a
glance, and how a diagram of the model is generated.

## Scope

In scope:

- Building the top-down composition tree rooted at an entity, or a forest from the
  bounded contexts when no identifier is given.
- Rendering the tree as text, as JSON, or as Graphviz dot.
- Trimming the output by depth and by excluded kinds.
- Marking a repeated entity so a cycle does not expand forever.

Out of scope:

- Reporting the parents that link to an entity, which showing an entity (cap-0005)
  does; the graph shows composition downward only.

## Actors

- Authors, who graph a subtree to understand its shape or to generate a diagram.

## Invariants

- In dot output, each entity is drawn once however many entities link to it.
- A repeated entity on a path is marked and not expanded again, so a cycle terminates.

## Specifications

### Tree and graph construction

- The text and JSON forms build a tree whose nodes carry whether each reference
  resolved and whether it repeats an ancestor on the path.
- The dot form builds a graph in which each entity is a single node and each link is a
  directed edge, so an entity linked from several places is drawn once.
- `--depth N` limits how far the tree expands, and `--exclude <kinds>` omits entities
  of the named kinds, for example `cap graph --format dot --exclude verification,task`.

## ADRs

## Scenarios

## Verification

- query/dot_test.go
- query/options_test.go

## Tasks
