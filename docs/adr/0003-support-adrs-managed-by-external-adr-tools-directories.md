# 3. Support ADRs managed by external adr-tools directories

Date: 2026-06-30

## Status

Accepted

## Context

cap stores its entities as Markdown files grouped by kind beneath the model root,
including ADRs under `adrs/` with `adr-NNNN` identifiers derived from the filename.

Many projects already manage their architectural decision records with adr-tools,
which records the ADR directory in a `.adr-dir` config file (for example `doc/adr`)
and names files with a bare, zero-padded number and a slug, for example
`0001-record-architecture-decisions.md`. Requiring those projects to move or
duplicate their ADRs into cap's `adrs/` directory, and to rename them with an `adr-`
prefix, would fork the source of truth and break adr-tools' own commands, which renumber
and supersede records in place.

## Decision

When a `.adr-dir` config file is present at the model root or any parent directory,
cap reads ADRs from the directory it names instead of the internal `adrs/`
directory. The search walks up from the model root, matching how adr-tools locates
its own config from the working directory upward. Absent a config, a conventional
`doc/adr` directory (adr-tools' default location) is used when one exists, so a
project that follows the convention is recognised without configuration.

ADR identifiers are synthesised from the leading number of the adr-tools filename, so
`0001-record-architecture-decisions.md` becomes `adr-0001`. Existing references of the
form `adr-NNNN` in capability documents continue to resolve unchanged. A filename that
already carries an `adr-` prefix is also accepted, and files without a leading number
(such as a README) are ignored.

cap does not scaffold into an external ADR directory. `cap new adr` reports that
adr-tools manages those records and points to `adr new`, leaving creation, numbering,
and superseding to adr-tools.

## Consequences

A project keeps a single source of truth for its ADRs and uses adr-tools' own
workflow to manage them, while cap reads, lists, links, and validates those ADRs
alongside the rest of the model.

The two naming conventions coexist: cap's `adr-NNNN` files in an internal `adrs/`
directory when no config is present, and adr-tools' bare-number files in the external
directory when one is. The internal `adrs/` directory is ignored while an external
directory is configured, so a project should keep its ADRs in one place rather than
both.

Identifiers in the external directory are derived from the file number rather than
written in the file, so two ADRs sharing a number (which adr-tools does not produce)
would collide; cap reports this as a duplicate-identifier error.
