# cap

`cap` documents a software system's functionality and behaviour, and reports on it.

It is two things:

- A standard for documenting a system as structured Markdown files in a simple
  folder layout.
- A set of tools that produce reports and visualisations from those files to guide
  development.

A bounded context groups the capabilities of part of the domain, each capability
holds the invariants it must uphold, and a specification records the design that
upholds them.

## Usage

Start from one entity, generate its template, then work outward to the entities it
relates to.

```bash
cap init                                  # create the folder layout and templates
cap new context "Policy enforcement"      # scaffold an entity, ready to edit
cap new concept "Policy"                  # name a thing the context operates on
cap new capability "Evaluate policies"
```

Each `cap new` writes a Markdown file with a heading for each section the entity
needs. Edit the file to fill in the detail, and reference related entities by
identifier under the relevant heading (for example a capability lists its context
under `## Metadata` and its invariants under `## Invariants`).

The tools report on the result:

| Command | Reports |
| --- | --- |
| `cap validate` | structural and reference problems, and gaps such as a capability with no verification or no specification |
| `cap list <kind>` | the entities of a kind, by identifier |
| `cap show <id>` | one entity, what it links to, and what links to it |
| `cap graph [id]` | the hierarchy beneath an entity (a context, its capabilities, their invariants), or the whole model from its bounded contexts when no identifier is given |
| `cap context <cap>` | the bundle an agent reads to generate code or tests for a capability |
| `cap review <id>` | a checklist for an agent to assess an entity against the conventions |
| `cap version` | the build version |

Every report accepts `--format json`.

Every command reads the model from the `./cap` directory under the current working
directory by default. Point it at a different layout with `--root <dir>`, or set the
`CAP_ROOT` environment variable.

An agent uses `cap context` and `cap review` to generate code and tests from a
capability and its invariants, and to fill documentation gaps by comparing the
source against the model.

## Entities

| Entity | Records |
| --- | --- |
| Bounded context | a part of the domain, grouping related capabilities and concepts |
| Concept | a thing in the context's language that capabilities operate on, such as a "policy" or a "decision" |
| Capability | an ability the system has, such as that it can "evaluate policies" |
| Invariant | a rule that must always hold, such as "an explicit deny overrides any permit" |
| Specification | the design of a capability or context that upholds a set of invariants |
| Verification | a test or procedure that proves the invariants hold |
| Scenario | a workflow that crosses several capabilities |
| ADR | an architectural decision that constrains how capabilities are implemented |
| Task | a unit of work |

An identifier is a lowercase prefix and a number: `ctx-0001`, `con-0001`, `cap-0003`,
`inv-0001`, `spec-0012`, `ver-0001`, `scn-0001`, `adr-0001`, `task-0341`.

A concept names a thing in the domain, a noun the capabilities act on. A capability
names an ability, a verb on one of those things. Defining the concepts of a context
keeps its capabilities atomic: one verb on one concept, rather than a broad
"manage policies" that hides several.

An invariant states a single rule. A specification describes the design that upholds
a set of those rules. A rule belongs in an invariant; the design that realises it
belongs in a specification.

### Relationships

A bounded context sits at the top of the hierarchy. It groups the concepts of its
domain and the capabilities that act on them. A concept belongs to one context. A
capability belongs to one context and links out to its invariants and verification. A
specification describes the design of a capability, or the design of a whole context
where the concern is cross-cutting. A scenario crosses several capabilities.

```text
Bounded context ──────────────────▶ Specification   a cross-cutting design
├─ Concept                           a thing in the context's language
└─ Capability ──▶ Invariant          a rule it must uphold
              ──▶ Specification      the design of this capability
              ──▶ Verification       evidence the invariants hold
              ──▶ ADR, Task          a decision, a unit of work

Scenario ──▶ Capability, Capability  a workflow across capabilities
```

A concept is referenced by name in the prose of the entities that use it. Tag the
name with the concept's identifier where it appears, for example "Policy (con-0001)",
so the reference can be traced.

A capability and an invariant may link to each other from either file, so a single
invariant constrains several capabilities. A specification is of a capability or of a
whole context.

## Getting started

Document a part of the system from the top down, starting with its bounded context.

```bash
# 1. Create the layout and templates.
cap init

# 2. Create the context, then a capability within it.
cap new context "Policy enforcement"
cap new capability "Evaluate policies"
```

`cap init` creates a directory per entity kind under the model root, alongside the
editable templates:

```text
cap/
├── contexts/
├── concepts/
├── capabilities/
├── invariants/
├── specifications/
├── scenarios/
├── verification/
├── adrs/
├── tasks/
└── .templates/
```

Each file is a structured Markdown document whose identifier is derived from its
filename, for example `capabilities/cap-0003-evaluate-policies.md`.

Edit the capability to record which context it belongs to, the rules it upholds, and
the design that upholds them. Invariants and specifications go inline; only an entity
shared across capabilities needs its own file.

```markdown
# Evaluate policies                            ← cap/capabilities/cap-0001-evaluate-policies.md

## Metadata
- context: ctx-0001

## Description
Decide whether an action is permitted, returning a permit or deny decision with the
reason for it.

## Scope
In scope:
- Combining a policy set and request inputs into a decision.
Out of scope:
- Authoring or storing policy definitions.

## Invariants
- Policies are evaluated consistently.
- An explicit deny overrides any permit.

## Specifications
### Policy evaluation design
- Policies compile to a decision tree, evaluated in a single pass.
- Deny-overrides is enforced by ordering deny nodes ahead of permit nodes.

## Verification
- internal/policy/eval_test.go
```

Check and explore the result:

```bash
cap validate              # structure, references, and gaps (untested, undocumented)
cap validate --strict     # exit non-zero on warnings as well as errors, for CI
cap graph                 # the whole model, from its bounded contexts down
cap graph ctx-0001        # the context, its capabilities, and their invariants
```

`cap validate` reports gaps as warnings and exits zero; `--strict` makes warnings
fail the exit status too, so a continuous integration job rejects an incomplete model.

## Templates

`cap init` installs a template for each entity kind into `cap/.templates/`. A
template is an ordinary entity document with guidance under each heading, and it is
also the schema: `cap validate` treats every heading as a required section unless its
title ends with `(optional)`. Edit the installed templates to change the headings
`cap new` scaffolds and the sections `cap validate` enforces; `cap init` preserves
templates that already exist, so customisations survive being run again.

## Development

Tools come from the Nix dev shell.

```bash
nix develop
```

## Tasks

### build

```bash
go build -o cap ./cmd/cap
```

### test

```bash
go test ./... -coverprofile=coverage.out
```

### fmt

```bash
go fmt ./...
```

### vulncheck

Report known vulnerabilities in the dependencies and in the code that calls them.

```bash
govulncheck ./...
```

### adr

Create an architectural decision record.

```bash
adr new "Title of the decision"
```

### release-snapshot

Build release artifacts locally without publishing, to check the GoReleaser
configuration.

```bash
goreleaser release --snapshot --clean
```

### release

Bump `.version` based on the commit count, then push the matching git tag. The
`release` GitHub Actions workflow builds and publishes the release from the tag.

```bash
version set
git add .version
git commit -m "chore: release $(version get)"
version push
```
