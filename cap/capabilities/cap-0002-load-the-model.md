# Load the model

## Metadata

- context: ctx-0001
- status: done

## Description

Read every Entity (con-0001) beneath a root directory and build the Model (con-0004):
the entities indexed by Identifier (con-0002), the map from each identifier to its
source file, and the list of problems found while parsing. Loading is the foundation
every reporting capability is built on. It tolerates malformed content, so a partial
model loads even when some files have problems.

## Scope

In scope:

- Reading each kind's directory and parsing each file into a typed entity.
- Resolving the identifier of each entity from its file name.
- Recording a problem for each malformed file or unparseable section.
- Loading ADRs from an external adr-tools directory when one is configured.

Out of scope:

- Checking references between entities, which validation (cap-0003) does.
- Writing to the model, which scaffolding (cap-0001) does.

## Actors

- Every reporting capability, which loads the model before producing its report.

## Invariants

- inv-0002
- inv-0004

## Specifications

### Loading and problem collection

- Each kind has a directory beneath the root; ADRs load from the `.adr-dir`
  directory instead when adr-tools configuration is present.
- Parsing extracts the title from the first heading and the items beneath each
  section heading, skipping code fences and prose outside sections.
- Inline invariants, specifications, and verification declared within a capability are
  given synthesised identifiers of the form `cap-0003/inv-1` and marked with the
  capability that owns them.
- A problem carries the file, line, severity, and message, so a report can list every
  problem sorted by location without stopping at the first.

## ADRs

- adr-0003

## Scenarios

## Verification

- store/store_test.go
- store/adrdir_test.go
- markdown/markdown_test.go

## Tasks
