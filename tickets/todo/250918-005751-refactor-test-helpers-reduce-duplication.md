---
priority: 2
description: "Refactor test helpers to reduce duplication while maintaining unit/integration separation"
created_at: "2025-09-18T00:57:51+09:00"
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

### Phase 1: Audit & Enhance Existing testutil
1. Review what `internal/testutil` already provides:
   - ✅ Git setup (`git.go` - `SetupGitRepo()`)
   - ✅ Ticket fixtures (`fixtures.go` - `TicketFixture()`)
   - ✅ Mock helpers (`mocks.go`)
   - ❌ Missing: Common test constants
2. Add missing functionality:
   - Add `constants.go` for shared test constants
   - Enhance git helpers if needed
   - Add any missing ticket creation patterns

### Phase 2: Migrate Duplicated Functions
1. **Replace `setupTestRepo` duplicates**:
   - `cmd/ticketflow/test_helpers.go` → use `testutil.SetupGitRepo()`
   - `test/integration/workflow_test.go` → use `testutil.SetupGitRepo()`
2. **Migrate ticket creation**:
   - Replace manual ticket creation with `testutil.TicketFixture()`
3. **Extract constants**:
   - Move test constants to `testutil/constants.go`

### Phase 3: Update Test Files (Gradual)
1. Start with files that have most duplication
2. Keep existing helpers as thin wrappers initially
3. Update imports to use `internal/testutil`
4. Run tests after each migration to ensure nothing breaks

### Phase 4: Documentation & Guidelines
1. Update `internal/testutil/README.md` with migration examples
2. Create `docs/testing-guidelines.md`
3. Document when to use testutil vs testharness
4. Add deprecation comments to old helpers

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
- [ ] All existing tests still pass
- [ ] Git setup duplication eliminated (setupTestRepo functions removed)
- [ ] Test constants centralized in testutil/constants.go
- [ ] testutil README updated with migration guide
- [ ] At least 2 existing test files migrated as examples
- [ ] Usage of testutil increased from 1 file to 10+ files