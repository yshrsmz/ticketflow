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

## Current Import Cycle Analysis

The import cycle that `testsupport` currently solves:
- `internal/git/*_test.go` → `internal/testutil` → `internal/mocks` → `internal/git` (CYCLE!)
- Current solution: `internal/git/*_test.go` → `internal/testsupport/gitconfig` (no cycle)

Files currently using `testsupport/gitconfig` (10 files total):
- `internal/git/` test files (git_test.go, worktree_test.go, git_context_test.go, git_divergence_test.go)
- `internal/cli/` test helpers and test files
- `internal/testutil/git.go` (already wrapping the functionality)

## Proposed Solutions

### Option 1: Interface in testutil/interfaces
Move just the `GitExecutor` interface to `internal/testutil/interfaces/executor.go`. Then `gitconfig.Apply` can live in `internal/testutil/git.go`. Both `internal/git` and `internal/mocks` can import the interface without creating a cycle.

**Pros:**
- Clear separation of interfaces from implementations
- Follows dependency inversion principle

**Cons:**
- Adds another package/directory
- May be overkill for a single interface

### Option 2: Extract mocks to separate package
Move `internal/mocks` out of the testutil import chain by keeping it as a standalone package. Then `internal/testutil` won't import `internal/mocks`, breaking the cycle. Test files that need both can import both separately.

**Pros:**
- Mocks remain isolated and reusable
- Clear separation of concerns

**Cons:**
- Tests need to import two packages instead of one convenient testutil
- May reduce discoverability of mock helpers

### Option 3: Inline GitExecutor interface
Define the `GitExecutor` interface in both `internal/git` and `internal/testutil` (interface segregation). Go allows this - interfaces are implicitly satisfied. Each package defines only what it needs.

**Pros:**
- No new packages needed
- Each package is self-contained
- Standard Go pattern

**Cons:**
- Slight duplication of interface definition
- Need to ensure interfaces stay compatible

### Option 4: Create testutil/gitconfig subpackage (RECOMMENDED)
Move gitconfig functionality to `internal/testutil/gitconfig` as a subpackage. This subpackage imports nothing from testutil, avoiding the cycle while keeping everything organized under testutil.

**Pros:**
- Keeps everything under testutil (achieves consolidation goal)
- Solves the import cycle cleanly
- Minimal code changes required
- Clear, maintainable structure

**Cons:**
- Still technically two packages, but better organized
- Slightly longer import path

### Option 5: Use build tags
Create `internal/testutil/git_config.go` with no imports to mocks, and `internal/testutil/git_config_with_mocks.go` with a `//go:build withmocks` tag. Git tests use the basic version, other tests use the full version.

**Pros:**
- Everything in one package
- Flexible based on build requirements

**Cons:**
- Build tags can be confusing
- May complicate the build process
- Not a common pattern for solving import cycles

## Notes

- Keep an eye on integration tests and CLI helpers that currently import the testsupport package; they should migrate transparently once the helper moves.
- Consider adding a lightweight dependency diagram to justify the restructuring if the solution introduces new interfaces.
- The previous ticket (250918-005751) created `testsupport` as a temporary measure; this ticket completes the consolidation effort.
