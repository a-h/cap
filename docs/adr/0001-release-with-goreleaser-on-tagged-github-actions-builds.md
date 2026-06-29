# 1. Release with GoReleaser on tagged GitHub Actions builds

Date: 2026-06-29

## Status

Accepted

## Context

`cap` is a single Go binary distributed from a public GitHub repository. Users on
Linux, macOS, and Windows need pre-built binaries for amd64 and arm64 without a Go
toolchain or a Nix installation. The project already pins its toolchain with a Nix
flake and runs tasks through `xc`, and it tracks the release version in a `.version`
file managed by the `version` tool (`github.com/a-h/version`).

A release process must:

- build cross-platform binaries with the build version embedded;
- gate every release on the test suite and a `nix build`, so a tag never ships a
  tree that fails continuous integration;
- run the same tools locally and in continuous integration, sourced from the flake.

## Decision

Releases are cut by pushing a `v`-prefixed tag, which the `version push` command
creates from the `.version` file. A GitHub Actions workflow (`release.yml`) triggers
on `v*` tags, installs Nix, runs `xc test` and `nix build` as a gate, then runs
`goreleaser release --clean` through the Nix dev shell. GoReleaser builds the
binaries, embeds the tag as `main.Version` via `ldflags`, and publishes an archive
per platform with a checksums file to a GitHub release.

A separate `ci.yml` workflow runs `xc test` and `nix build` on every push to `main`
and on pull requests, so the flake build is exercised continuously rather than only
at release time.

`goreleaser` is added to the flake `devTools` so the same version (from the pinned
nixpkgs) runs locally and in continuous integration.

Alternatives considered:

- A hand-written workflow that runs `go build` for each platform and uploads
  artifacts with the GitHub CLI. Rejected because GoReleaser already handles the
  build matrix, archives, checksums, and changelog, and is the established tool for
  Go release automation.
- Building release binaries from the Nix flake. Rejected because the flake build
  targets the host platform for continuous integration verification, whereas
  GoReleaser cross-compiles the full distribution matrix in one run. The flake build
  remains the gate that proves `nix build` works.

## Consequences

A release needs only a version bump and a tag push; the workflow produces signed
checksums and per-platform archives without manual steps. Every release and every
change to `main` proves that `nix build` works, satisfying the requirement that the
Nix build always builds.

Both the flake build and GoReleaser embed the build version through `ldflags`. The
flake reads it from `.version`, and GoReleaser reads it from the tag, so the same
`.version` file is the single source of the version across both paths.

GoReleaser configuration is duplicated knowledge alongside the flake's build inputs
(both name the module and entry point). The two are validated together by the
`release-snapshot` task and the continuous integration `nix build`.
