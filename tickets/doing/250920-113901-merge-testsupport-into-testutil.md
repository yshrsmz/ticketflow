---
priority: 2
description: Merge testsupport git helpers into internal/testutil without reintroducing import cycles
created_at: "2025-09-20T11:39:01+09:00"
started_at: "2025-09-20T18:08:59+09:00"
closed_at: null
related:
    - parent:250918-005751-refactor-test-helpers-reduce-duplication
---

# Merge testsupport Git Helpers into testutil

## Context

`internal/testsupport/gitconfig` exists so git package tests can share `gitconfig.Apply` without importing `internal/testutil`, but this split creates confusing duplication between two helper packages. We now need to restructure the helpers so the git configuration logic can live under `internal/testutil` while keeping the import graph acyclic.

## Goals

- Collapse the standalone testsupport package into `internal/testutil` and expose a single git configuration helper.
- Break or avoid the current import cycle (`internal/git` → `internal/testutil` → `internal/mocks` → `internal/git`) so git tests can consume the shared helper safely.
- Update documentation and call sites so future contributors know which helper to use.

## Tasks

- [ ] Map the current import relationships between `internal/git`, `internal/testutil`, and `internal/testsupport`.
- [ ] Prototype a dependency break (e.g. move mocks or introduce interfaces) that allows `gitconfig.Apply` to live under `internal/testutil`.
- [ ] Move the helper, delete `internal/testsupport/gitconfig`, and update all call sites.
- [ ] Refresh `internal/testutil/README.md` (and other docs if needed) with the new structure.
- [ ] Run `make fmt`, `make vet`, and `make lint`.
- [ ] Run `make test`.
- [ ] Capture before/after notes in this ticket and seek developer approval before closing.

## Notes

- Keep an eye on integration tests and CLI helpers that currently import the testsupport package; they should migrate transparently once the helper moves.
- Consider adding a lightweight dependency diagram to justify the restructuring if the solution introduces new interfaces.
