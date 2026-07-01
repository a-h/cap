# Validate the model

## Metadata

- context: ctx-0002
- status: done

## Description

Check a loaded Model (con-0004) for structural and reference problems, and for gaps
that weaken traceability, and report each as a problem with its file and line. This
is how an author learns that a Reference (con-0003) is dangling, that a link is
declared on only one side, or that a capability has no verification, before those
gaps mislead a later report.

## Scope

In scope:

- Checking that every reference names an existing entity of the right kind.
- Checking that a many-to-many link is declared on both entities.
- Reporting gaps: a capability with no verification, no specification, or no status.
- Warning when the same inline invariant text appears in several capabilities.

Out of scope:

- Parsing the model, which loading (cap-0002) does; validation reports the parse
  problems loading already found alongside its own.
- Judging the prose of an entity against the conventions, which review (cap-0008)
  supports.

## Actors

- Authors, who run validation to find problems and gaps before committing.
- Continuous integration, which fails a build when validation reports errors.

## Invariants

- inv-0001
- inv-0003

## Specifications

### Checks and severities

- Reference integrity, symmetric links, and coverage gaps are reported as warnings;
  the parse problems loading found may be errors.
- A `--strict` mode treats warnings as errors, so a pipeline can require a clean model.
- Problems are sorted by file and line, and the exit code reflects the number of
  errors.

## ADRs

## Scenarios

- scn-0002

## Verification

- validate/validate_test.go

## Tasks
