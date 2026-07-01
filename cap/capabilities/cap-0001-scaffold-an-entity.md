# Scaffold an entity

## Metadata

- context: ctx-0001
- status: done

## Description

Create a new Entity (con-0001) from the template for its kind, ready for an author to
edit. Given a kind and a name, the capability allocates the next free identifier
(con-0002) for that kind, writes a Markdown file whose heading is the name and whose
sections come from the template, and reports the path written. It also creates the
directory layout and installs the templates when a model is first set up.

## Scope

In scope:

- Allocating the next free identifier for a kind.
- Rendering the template for a kind into a named file in the correct directory.
- Creating the directory layout and installing default templates.

Out of scope:

- Filling in the content of an entity, which the author does.
- Creating ADRs when an external adr-tools directory manages them; those are created
  with `adr new`.

## Actors

- Authors documenting a system, who scaffold an entity before editing it.

## Invariants

- inv-0002

## Specifications

### Identifier allocation

- The next free number for a kind is found by scanning the kind's directory for
  existing identifiers with the same prefix and taking one past the highest.
- The file name is the identifier followed by a slug of the name, for example
  `cap-0001-scaffold-an-entity.md`.
- The template's `(optional)` section markers are stripped as the file is rendered, so
  a scaffolded entity shows every section without the marker.

## ADRs

## Scenarios

- scn-0002

## Verification

- store/scaffold_test.go
- template/template_test.go

## Tasks
