---
priority: 3
description: "Phase 3: Add comprehensive testing for interspersed flag support"
created_at: "2025-09-26T17:23:25+09:00"
started_at: null
closed_at: null
related:
  - "parent:250924-143504-migrate-to-pflag-for-flexible-cli-args"
  - "blocked-by:250926-165945-phase1-pflag-basic-import-migration"
  - "blocked-by:250926-172234-phase2-refactor-flag-helpers-for-pflag"
---

# Phase 3: Comprehensive Interspersed Flag Testing

## Objective
Add comprehensive test coverage to verify interspersed flag functionality works correctly and update existing tests that validate argument ordering.

## Prerequisites
- Phase 1 completed (pflag imports)
- Phase 2 completed (flag helper refactoring)

## Scope
- Add integration tests for interspersed flag support
- Update existing tests that expect "unexpected arguments" errors
- Create test utilities for flag parsing scenarios
- Document expected behavior

## Important Phase 1 Findings

During Phase 1 migration, we discovered pflag-specific behaviors that need testing:
- Single-character flags registered with `StringVar`/`IntVar` don't work as shorthand flags
- pflag requires `StringVarP`/`BoolVarP` for proper short/long flag registration
- Test cases had to be updated from `-s` to `--s` for single-char StringVar registrations
- The reflection workaround in Phase 1's flag_types.go will be removed in Phase 2

## Test Categories

### 1. Update Existing Validation Tests

Several commands have tests that expect errors for flags after positional args:
```go
// Current test (will fail after pflag migration)
func TestShowCommand_Validate_ExtraArgs(t *testing.T) {
    cmd := NewShowCommand()
    err := cmd.Validate(flags, []string{"ticket-123", "--format", "json"})
    assert.Error(t, err) // This will fail - no longer an error!
}
```

These tests need updating in:
- `show_test.go` - Line 61: "unexpected arguments after ticket ID"
- `start_test.go` - Line 63: "unexpected arguments after ticket ID"
- `close_test.go` - Similar validation
- `cleanup_test.go` - Similar validation

### 2. Add Integration Tests

Create `internal/cli/commands/interspersed_flags_integration_test.go`:

```go
package commands_test

import (
    "context"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

func TestInterspersedFlags_ShowCommand(t *testing.T) {
    tests := []struct {
        name     string
        args     []string
        wantErr  bool
    }{
        {
            name: "flags before positional args",
            args: []string{"--format", "json", "ticket-123"},
        },
        {
            name: "flags after positional args",
            args: []string{"ticket-123", "--format", "json"},
        },
        {
            name: "flags interspersed",
            args: []string{"--format", "json", "ticket-123", "--verbose"},
        },
        {
            name: "short flags after args",
            args: []string{"ticket-123", "-o", "json"},
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Test implementation
        })
    }
}
```

### 3. Test All Affected Commands

Commands that take positional arguments and need testing:
- `show <ticket-id>`
- `start <ticket-id>`
- `close [ticket-id]`
- `cleanup <ticket-id>`
- `restore <ticket-id>`
- `worktree clean`
- `new <slug>`

### 4. Edge Case Testing

```go
func TestInterspersedFlags_EdgeCases(t *testing.T) {
    tests := []struct {
        name string
        args []string
    }{
        {
            name: "double dash terminator",
            args: []string{"--format", "json", "--", "ticket-123", "--not-a-flag"},
        },
        {
            name: "multiple positional args with flags",
            args: []string{"arg1", "--flag", "value", "arg2", "--another", "arg3"},
        },
        {
            name: "boolean flags interspersed",
            args: []string{"ticket-123", "--force", "--verbose"},
        },
    }
}
```

### 5. Create Test Helper

Add to `internal/cli/commands/testutils.go`:
```go
// ParseArgsWithCommand simulates real CLI flag parsing
func ParseArgsWithCommand(t *testing.T, cmd Command, args []string) (interface{}, []string, error) {
    fs := flag.NewFlagSet(cmd.Name(), flag.ContinueOnError)
    flags := cmd.SetupFlags(fs)

    err := fs.Parse(args)
    if err != nil {
        return nil, nil, err
    }

    return flags, fs.Args(), nil
}
```

## Testing Strategy

### Manual Testing Script
Create `scripts/test-interspersed-flags.sh`:
```bash
#!/bin/bash

echo "Testing interspersed flags..."

# Test show command
./ticketflow show ticket-123 --format json || exit 1
./ticketflow show --format json ticket-123 || exit 1

# Test start command
./ticketflow start ticket-456 --force || exit 1
./ticketflow start --force ticket-456 || exit 1

# Test with short flags
./ticketflow show ticket-789 -o json || exit 1

echo "All interspersed flag tests passed!"
```

### Regression Testing
Ensure backward compatibility:
1. All existing command patterns still work
2. Help text displays correctly
3. Error messages remain clear
4. Flag precedence (if both short/long provided) works

## Documentation Updates

Update documentation to show new capability:
- README examples
- CLAUDE.md AI integration section
- Help text examples

## Success Criteria
- All existing tests updated and passing
- New integration tests covering interspersed scenarios
- Manual testing script validates real CLI behavior
- No regression in existing functionality
- Documentation reflects new capabilities

## Estimated Effort
- 1 hour for test updates and new integration tests
- 30 minutes for manual testing and documentation

## Notes
- Focus on commands that AI agents commonly use
- Ensure error messages are still helpful when actual errors occur
- Consider adding benchmarks to ensure no performance regression