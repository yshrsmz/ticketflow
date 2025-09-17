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

The main duplication is in git repository setup code (~200 lines) and test constants.

## Options Considered

### Option 1: Minimal Consolidation
Keep existing structure but extract only common constants and utilities.
- ✅ Minimal disruption
- ❌ Doesn't address main duplication

### Option 2: Unified Test Helpers Package ✅ **CHOSEN**
Create central `internal/testutil` package with sub-packages for organization.
- ✅ Addresses real duplication
- ✅ Maintains unit/integration separation
- ✅ Gradual migration possible
- ✅ Simple and pragmatic

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

1. **Respects Testing Philosophy**: Maintains clear separation between unit tests (mocks) and integration tests (real environment)
2. **Addresses Real Problem**: Directly targets actual duplication (git setup, constants) without over-engineering
3. **Pragmatic and Simple**: Organizes helpers better without creating unnecessary abstractions
4. **Gradual Migration**: Allows incremental updates without breaking existing tests
5. **Proven Pattern**: Similar structure used successfully in many Go projects

## Implementation Plan

### Phase 1: Create Structure
```
internal/testutil/
├── constants.go           # Shared test constants
├── unit/
│   └── fixture.go         # Mock-based fixtures
├── integration/
│   ├── repo.go           # Git repository setup
│   └── ticket.go         # Ticket creation helpers
└── builders/
    └── ticket.go         # Test data builders (if needed)
```

### Phase 2: Extract Common Code
1. Move shared constants to `testutil/constants.go`
2. Create unified git setup in `testutil/integration/repo.go`
3. Extract common ticket creation patterns

### Phase 3: Update Existing Code
1. Update `cmd/ticketflow/test_helpers.go` to use testutil
2. Remove duplicate in `test/integration/workflow_test.go`
3. Keep existing helpers as thin wrappers for compatibility

### Phase 4: Documentation
1. Create `docs/testing-guidelines.md`
2. Document when to use each helper
3. Add examples of proper usage

## Expected Benefits
- Eliminate ~200 lines of duplicated code
- Single source of truth for test constants
- Clearer test organization
- Easier to maintain and extend
- Better onboarding for new developers

## Migration Strategy
- Keep existing tests working (backward compatibility)
- Update gradually as files are touched
- New tests use the new structure
- Full migration over several months

## Success Criteria
- [ ] All existing tests still pass
- [ ] Git setup duplication eliminated
- [ ] Test constants centralized
- [ ] Documentation created
- [ ] At least 2 existing test files migrated as examples