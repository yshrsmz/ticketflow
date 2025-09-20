---
priority: 2
description: Refactor test helpers to reduce duplication while maintaining unit/integration separation
created_at: "2025-09-18T00:57:51+09:00"
started_at: "2025-09-20T01:33:29+09:00"
closed_at: null
---

# Refactor Test Helpers to Reduce Duplication

## Problem Statement

The codebase has multiple test helper files with duplicated functionality across different packages:
- `cmd/ticketflow/test_helpers.go` - CLI handler test helpers
- `internal/cli/test_helpers.go` - CLI app test helpers with mocks
- `internal/cli/commands/test_helpers.go` - Command test helpers with mocks
- `internal/cli/commands/testharness/harness.go` - Integration test environment
- `test/integration/workflow_test.go` - Has its own setupTestRepo (duplicate)

**Critical Discovery**: `internal/testutil` package already exists with comprehensive test utilities (git setup, fixtures, mocks, etc.) but is only used by 1 test file. The package was created but migration never happened, resulting in parallel test utility systems.

### Current State Analysis

| Package/File | Purpose | Usage | Key Functions |
|-------------|---------|--------|---------------|
| `internal/testutil` | Comprehensive utilities | 1 file only | SetupGitRepo(), TicketFixture(), mocks |
| `cmd/ticketflow/test_helpers.go` | CLI handler tests | ~5 files | setupTestRepo() variants |
| `internal/cli/test_helpers.go` | CLI app unit tests | ~10 files | testFixture, mocks |
| `internal/cli/commands/test_helpers.go` | Command unit tests | ~10 files | TestFixture, mocks |
| `testharness/harness.go` | Integration tests | ~15 files | TestEnvironment (keep as-is) |
| `test/integration/workflow_test.go` | Integration test | 1 file | Duplicate setupTestRepo() |

The main duplication is in git repository setup code (~200 lines) and test constants.

### Additional Findings (2025-09-19)

- `internal/testutil.SetupGitRepo` currently leaves repositories on whatever default branch the local git prefers and does not disable commit signing, so it cannot replace the bespoke helpers that guarantee a `main` branch and predictable config (`internal/testutil/git.go:37-57`, `cmd/ticketflow/test_helpers.go:44-130`).
- `internal/testutil.CreateConfigFile` writes hard-coded values that break our expected directory layout (e.g., it emits `todo_dir: tickets/todo` instead of `todo`) and ignores the provided `config.Config`, making the existing options like `WithCustomConfig` ineffective (`internal/testutil/filesystem.go:60-103`, `internal/testutil/filesystem.go:217-243`).
- The scaffold does not create the `tickets/.current` marker that several CLI handler tests assume exists (`cmd/ticketflow/test_helpers.go:117-119`).
- `internal/cli/commands/test_helpers.go` appears to be dead code today—`rg` finds no call sites for `NewTestFixture`, `SetupMockForStart`, etc.—so we should decide whether to remove it or fold its intent into `testutil` instead of migrating it as-is.

## Options Considered

### Option 1: Minimal Consolidation
Keep existing structure but extract only common constants and utilities.
- ✅ Minimal disruption
- ❌ Doesn't address main duplication

### Option 2: Migrate to Existing internal/testutil ✅ **CHOSEN**
Use and enhance the existing `internal/testutil` package instead of creating new structure.
- ✅ Package already exists with good design
- ✅ Addresses real duplication
- ✅ Maintains unit/integration separation
- ✅ Gradual migration possible
- ✅ Leverages existing work

### Option 3: TestHarness Everywhere
Expand testharness to handle both unit and integration tests.
- ❌ Over-engineering for unit tests
- ❌ Violates unit testing principles
- ❌ Makes unit tests slow

### Option 4: Domain-Specific Helpers
Keep separate helpers but make them domain-focused.
- ✅ Logical separation
- ❌ Doesn't fix git setup duplication

### Option 5: Test Builder Pattern
Implement builder pattern for test setup.
- ✅ Flexible and composable
- ❌ More complex than needed
- ❌ Over-engineering for current needs

### Option 6: Keep As-Is
Accept current duplication as legitimate.
- ✅ No risk
- ❌ Duplication continues to grow

### Option 7: Extract Only Git Setup
Minimal refactor focusing on most duplicated part.
- ✅ Low risk
- ❌ Doesn't address other duplication

### Option 8: Test Traits System
Create composable test traits.
- ✅ Modern approach
- ❌ Over-complex for Go
- ❌ Team unfamiliarity

## Why Option 2 Was Chosen

1. **Leverages Existing Work**: `internal/testutil` already exists with well-designed utilities
2. **Respects Testing Philosophy**: Maintains clear separation between unit tests (mocks) and integration tests (real environment)
3. **Addresses Real Problem**: Directly targets actual duplication (git setup, constants) without over-engineering
4. **Pragmatic and Simple**: Uses existing package rather than creating new abstractions
5. **Gradual Migration**: Allows incremental updates without breaking existing tests
6. **Already Documented**: Package has comprehensive README and examples

## Implementation Plan

### Phase 1: Harden `internal/testutil`
1. ✅ Refactor `SetupGitRepo` to delegate to `SetupGitRepoWithOptions` so the default branch is forced to `main`, empty commits are optional, and commit signing is disabled for tests.
2. ✅ Teach the filesystem scaffold to honour caller-provided configs: generate `.ticketflow.yaml` from the struct, fix the ticket directory fields (`todo`, `doing`, `done`), and make `WithCustomConfig`/`Without*` options actually influence the output.
3. ✅ Add a helper (e.g., `SetupTicketflowRepo`) that bundles git init, config creation, ticket directories, and the `tickets/.current` marker so packages no longer need to hand-roll the same steps.
4. ✅ Introduce `constants.go` once the shared helpers are in place, limiting it to truly global fixtures (ticket IDs, timestamps) that multiple packages rely on.

### Phase 2: Introduce adapter wrappers
1. ✅ Update `cmd/ticketflow/test_helpers.go` and the integration test package to call the new scaffold while keeping their existing signatures, easing migration.
2. ✅ Decide the fate of `internal/cli/commands/test_helpers.go` (delete if unused, otherwise port the useful pieces onto the shared helpers).

### Phase 3: Migrate CLI/unit tests
1. Remove direct git/config setup from `cmd/ticketflow/test_helpers.go`, replacing it with the shared helper.
2. Swap duplicated constants/usages in `cmd/ticketflow` and `internal/cli` tests to reference `testutil/constants.go`.
3. Run focused suites (`go test ./cmd/ticketflow/...`, `go test ./internal/cli/...`) after each migration step.

### Phase 4: Migrate integration tests
1. Replace the `setupTestRepo` implementations under `test/integration` with the shared helper while keeping `os.Chdir` discipline intact.
2. Re-run the integration suite (`make test-integration`) to confirm behaviours like branch naming and current-ticket symlink handling remain intact.

### Phase 5: Documentation & cleanup
1. Expand `internal/testutil/README.md` with the new helpers, options, and migration examples.
2. Mark any remaining ad-hoc helpers with deprecation comments and open follow-up tickets for packages that still need to migrate.

## Expected Benefits
- Eliminate ~200 lines of duplicated code
- Single source of truth for test constants
- Clearer test organization
- Easier to maintain and extend
- Better onboarding for new developers

## Migration Strategy
- Keep existing tests working (backward compatibility)
- Update gradually as files are touched
- New tests should use `internal/testutil` utilities
- Full migration over several months
- Keep testharness for integration tests (it's working well)

## Success Criteria
- [x] `internal/testutil.SetupGitRepo` forces a `main` branch, disables signing, and is covered by unit tests.
- [x] The shared scaffold can emit the canonical `.ticketflow.yaml` (including configurable worktree settings) and create the `tickets/.current` marker.
- [x] `cmd/ticketflow/test_helpers.go` and `test/integration` no longer invoke `exec.Command` directly for repo setup (they rely on the shared helper/wrapper).
- [x] Shared test constants live in `internal/testutil/constants.go` with consumers updated to use them.
- [x] `internal/testutil/README.md` documents the new helpers and migration guidance.
- [x] `make test` passes locally after the migrations.

### Progress (2025-09-19)

- Hardened `internal/testutil` git and filesystem scaffolds with new unit coverage, added shared constants, switched CLI + integration repositories to the consolidated helper while deleting the unused commands helper, documented usage, and verified `make test` succeeds.
- Rewired every integration suite to consume the shared setup helper so `setupTestRepo` no longer shells out to git directly.
