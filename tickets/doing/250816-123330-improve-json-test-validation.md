---
priority: 2
description: Improve validation logic for JSON-based tests to ensure accuracy and reliability
created_at: "2025-08-16T12:33:30+09:00"
started_at: "2025-08-17T12:26:35+09:00"
closed_at: null
related:
    - parent:250815-175624-test-coverage-maintenance-commands
---

# Ticket Overview

Improve validation logic for JSON-based tests across the ticketflow codebase to ensure consistent, robust testing of JSON output. Currently, many integration tests use fragile string contains checks instead of properly parsing and validating JSON structure, which can lead to false positives and brittle tests.

## Problem Statement

After analyzing the codebase, I identified inconsistent JSON validation patterns:

1. **Fragile string validation** (found in multiple test files):
   - `status_integration_test.go`: Uses `assert.Contains` for JSON fragments
   - `worktree_integration_test.go`: Uses `assert.Contains` for JSON fragments
   - Several other integration tests follow this anti-pattern

2. **Proper JSON validation** (good examples to follow):
   - `cleanup_integration_test.go` (lines 93-105): Properly parses JSON and validates structure
   - Uses `json.Unmarshal` and type assertions to check actual structure

3. **Impact**: 
   - String contains checks don't validate proper JSON structure
   - Can pass even if JSON is malformed
   - Makes tests brittle and harder to maintain
   - Doesn't verify data types or nested structures

## Tasks

- [x] Analyze current JSON test validation patterns across the codebase
- [ ] Design a test helper for consistent JSON validation
- [ ] Create JSON validation helper in testharness package
- [ ] Update status_integration_test.go to use proper JSON validation
- [ ] Update worktree_integration_test.go to use proper JSON validation
- [ ] Update cleanup_integration_test.go to improve JSON validation consistency
- [ ] Update remaining integration tests with JSON output to use helper
- [ ] Run `make test` to ensure all tests pass
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update ticket with implementation insights
- [ ] Get developer approval before closing

## Implementation Plan

### 1. Create JSON Validation Helper

Add to `internal/cli/commands/testharness/helpers.go`:
```go
// ValidateJSON parses JSON output and returns the unmarshaled structure
func ValidateJSON(t *testing.T, output string) map[string]interface{} {
    // Strip any non-JSON content (status messages, etc.)
    // Parse JSON
    // Return structured data for assertions
}

// AssertJSONField validates a specific field in JSON output
func AssertJSONField(t *testing.T, jsonData map[string]interface{}, path string, expected interface{}) {
    // Navigate nested JSON structure
    // Assert field value matches expected
}
```

### 2. Files to Update

Integration test files that need JSON validation improvements:
- `status_integration_test.go` (lines 57-60)
- `worktree_integration_test.go` (lines 59-61)
- `worktree_list_integration_test.go`
- `start_integration_test.go`
- `show_integration_test.go`
- `restore_integration_test.go`
- `new_integration_test.go`
- `close_integration_test.go`
- `cleanup_integration_test.go` (improve consistency)

### 3. Example Transformation

**Before** (fragile):
```go
assert.Contains(t, outputStr, `"id": "current-ticket"`)
assert.Contains(t, outputStr, `"description": "Test ticket for JSON"`)
```

**After** (robust):
```go
jsonData := testharness.ValidateJSON(t, outputStr)
testharness.AssertJSONField(t, jsonData, "current_ticket.id", "current-ticket")
testharness.AssertJSONField(t, jsonData, "current_ticket.description", "Test ticket for JSON")
```

## Benefits

1. **Robust validation**: Actually parses JSON and validates structure
2. **Better error messages**: Clear indication of what field failed validation
3. **Type safety**: Can validate data types, not just string presence
4. **Maintainability**: Single helper to update if JSON format changes
5. **Consistency**: All tests use the same validation approach

## Notes

- This is a sub-ticket of the test coverage improvement initiative
- Focuses on improving existing tests rather than adding new ones
- Aligns with the project's goal of robust, maintainable test suite
- Part of the broader effort to improve test quality following integration testing patterns