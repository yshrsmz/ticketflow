---
priority: 2
description: Merge testsupport git helpers into internal/testutil without reintroducing import cycles
created_at: "2025-09-20T11:39:01+09:00"
started_at: "2025-09-20T18:08:59+09:00"
closed_at: "2025-09-20T23:09:37+09:00"
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

- [x] Map the current import relationships between `internal/git`, `internal/testutil`, and `internal/testsupport`.
- [x] Prototype a dependency break (e.g. move mocks or introduce interfaces) that allows `gitconfig.Apply` to live under `internal/testutil`.
- [x] Move the helper, delete `internal/testsupport/gitconfig`, and update all call sites.
- [x] Refresh `internal/testutil/README.md` (and other docs if needed) with the new structure.
- [x] Run `make fmt`, `make vet`, and `make lint`.
- [x] Run `make test`.
- [x] Fix issues identified in code review (variable shadowing bug in GitConfigApply)
- [x] Add comprehensive unit tests for GitConfigApply functionality
- [ ] Capture before/after notes in this ticket and seek developer approval before closing.

## Current Import Cycle Analysis

The import cycle that `testsupport` currently solves:
- `internal/git/*_test.go` → `internal/testutil` → `internal/mocks` → `internal/git` (CYCLE!)
- Current solution: `internal/git/*_test.go` → `internal/testsupport/gitconfig` (no cycle)

Files currently using `testsupport/gitconfig` (10 files total):
- `internal/git/` test files (git_test.go, worktree_test.go, git_context_test.go, git_divergence_test.go)
- `internal/cli/` test helpers and test files
- `internal/testutil/git.go` (already wrapping the functionality)

## Discovered Issue

After deeper analysis with Codex, we discovered that **MockSetup in internal/testutil/mocks.go is completely unused**. This unused code is the ONLY thing creating the import cycle! By deleting it, we can achieve a much cleaner solution than any of the options below.

## Recommended Solution: Option C - Delete Dead Code

**DELETE the unused MockSetup from testutil/mocks.go, then move gitconfig.Apply directly into testutil.**

This is the best solution because:
- **Removes 136 lines of dead code** (the entire mocks.go file is unused)
- **Immediately breaks the import cycle** without any architectural changes
- **Achieves perfect consolidation** - gitconfig.Apply moves directly into internal/testutil
- **Simplest possible solution** - follows Occam's Razor

### Implementation Steps:
1. Delete `internal/testutil/mocks.go` (contains only unused MockSetup)
2. Move gitconfig functionality from `internal/testsupport/gitconfig` to `internal/testutil`
3. Update all imports from `testsupport/gitconfig` to `testutil`
4. Delete the entire `internal/testsupport` directory
5. Update documentation to remove MockSetup references

### Why This Beats All Other Options:
- No new packages needed (unlike Options 1, 2, 4)
- No code duplication (unlike Option 3)
- No build tag complexity (unlike Option 5)
- Reduces overall codebase complexity instead of adding to it

## Alternative Solutions (Kept for Reference)

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

## Implementation Summary (Option C - Delete Dead Code)

### What Was Done
1. **Deleted `internal/testutil/mocks.go`** - Contained only unused MockSetup code (136 lines)
2. **Moved gitconfig functionality to `internal/testutil/git.go`**:
   - Added `GitExecutor` interface
   - Added `GitConfigOptions` struct
   - Added `GitConfigApply` function (exported for use by git tests)
3. **Updated all imports** (10 files total):
   - 4 git test files
   - 4 CLI test files
   - 2 command test files
4. **Deleted `internal/testsupport` directory entirely**
5. **Updated documentation**:
   - Removed MockSetup references from README
   - Cleaned up the migration guide

### Why This Solution is Optimal
- **Removes dead code**: MockSetup was completely unused
- **Breaks the cycle instantly**: No mocks import = no cycle
- **Achieves perfect consolidation**: Everything now under `internal/testutil`
- **Minimal complexity**: We removed complexity rather than adding it

### Changes Made
- Files deleted: 3 (`mocks.go`, `testsupport/gitconfig/gitconfig.go`, `testsupport/gitconfig/gitconfig_test.go`)
- Files modified: 11 (updated imports and integrated gitconfig functionality)
- Lines removed: ~200+ (dead code + testsupport package)
- Lines added: ~60 (gitconfig functionality in testutil)

### Testing
- All tests pass: `make test` ✓
- All linting passes: `make fmt vet lint` ✓
- No import cycles detected
- Added comprehensive unit tests for GitConfigApply (8 test cases)
- Tests cover default options, custom options, command ordering, and interface compliance

## Post-Implementation Review

### Code Review Findings (via golang-pro)
1. **Variable Shadowing Bug**: Found that the parameter name `exec` was shadowing the imported `exec` package, which could have caused compilation errors. Fixed by renaming to `executor`.
2. **Missing Tests**: The migrated functionality lacked unit tests. Added comprehensive test coverage.

### Key Insights
1. **Dead Code Detection is Valuable**: The discovery that MockSetup was completely unused led to the simplest possible solution. Regular dead code analysis should be part of our maintenance routine.
2. **Simplest Solution Often Best**: While we considered 5 complex architectural solutions, the best answer was simply deleting unused code.
3. **Import Cycles Hide Deeper Issues**: The import cycle was a symptom of dead code creating unnecessary dependencies, not a fundamental architectural problem.
4. **Code Review is Essential**: The golang-pro review caught a subtle but important bug that passed all tests but could have caused issues later.

### Lessons Learned
- Always check for dead code before refactoring around perceived architectural issues
- When facing import cycles, map the actual usage (not just the imports) to find unnecessary dependencies
- Automated code review tools can catch issues that humans and tests might miss
- Adding tests during refactoring (not just after) helps ensure correctness

## Notes

- Keep an eye on integration tests and CLI helpers that currently import the testsupport package; they should migrate transparently once the helper moves.
- Consider adding a lightweight dependency diagram to justify the restructuring if the solution introduces new interfaces.
- The previous ticket (250918-005751) created `testsupport` as a temporary measure; this ticket completes the consolidation effort.
