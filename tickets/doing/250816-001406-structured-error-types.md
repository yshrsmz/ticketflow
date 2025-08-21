---
priority: 2
description: Refactor CLI commands to use structured error types for better error handling
created_at: "2025-08-16T00:14:06+09:00"
started_at: "2025-08-21T16:44:08+09:00"
closed_at: null
related:
    - parent:250815-171607-improve-command-test-coverage
---

# Refactor to Structured Error Types

## Overview

During PR review for test coverage improvements (#71), a suggestion was made to use structured error types instead of string errors. While we implemented this for one specific case (`CloseTicketInternalError`), analysis revealed a broader pattern of errors across all CLI commands that would benefit from structured types.

## Background

In PR #71 review comment, it was suggested to replace:
```go
return fmt.Errorf("internal error: closed ticket is nil (isCurrentTicket=%v)", isCurrentTicket)
```

With:
```go
return &CloseTicketInternalError{IsCurrentTicket: isCurrentTicket}
```

This led to discovering recurring error patterns across all commands that could benefit from similar treatment.

## Identified Error Patterns

Analysis found these common error types across CLI commands:

1. **Invalid flags type errors** (11 occurrences)
   - Pattern: `"invalid flags type: expected *%sFlags, got %T"`
   - Files: close.go, start.go, new.go, list.go, status.go, restore.go, show.go

2. **Invalid format errors** (8 occurrences)
   - Pattern: `"invalid format: %q (must be %q or %q)"`
   - Files: close.go, start.go, new.go, list.go, status.go, restore.go, show.go, worktree_list.go

3. **Unexpected arguments errors** (5 occurrences)
   - Pattern: `"unexpected arguments after X: %v"`
   - Files: close.go, start.go, new.go, show.go

4. **Missing argument errors** (2 occurrences)
   - Pattern: `"missing X argument"`
   - Files: start.go

## Proposed Solution

Create a common errors package for CLI commands with structured types:

```go
// internal/cli/commands/errors.go

type InvalidFlagsTypeError struct {
    Expected string
    Got      interface{}
}

func (e *InvalidFlagsTypeError) Error() string {
    return fmt.Sprintf("invalid flags type: expected %s, got %T", e.Expected, e.Got)
}

type InvalidFormatError struct {
    Given        string
    ValidFormats []string
}

func (e *InvalidFormatError) Error() string {
    return fmt.Sprintf("invalid format: %q (must be %s)", 
        e.Given, strings.Join(e.ValidFormats, " or "))
}

type UnexpectedArgumentsError struct {
    After string
    Args  []string
}

func (e *UnexpectedArgumentsError) Error() string {
    return fmt.Sprintf("unexpected arguments after %s: %v", e.After, e.Args)
}

type MissingArgumentError struct {
    ArgumentName string
}

func (e *MissingArgumentError) Error() string {
    return fmt.Sprintf("missing %s argument", e.ArgumentName)
}
```

## Benefits

1. **Type Safety**: Errors can be handled differently based on their type
2. **Consistency**: Standardized error messages across all commands
3. **Better Testing**: Can assert on error types, not just strings
4. **Future Extensibility**: Easy to add fields for additional context
5. **Structured Logging**: If we add structured logging later, these errors will provide better context

## Tasks

- [ ] Create `internal/cli/commands/errors.go` with common error types
- [ ] Refactor close.go to use structured errors
- [ ] Refactor start.go to use structured errors
- [ ] Refactor new.go to use structured errors
- [ ] Refactor list.go to use structured errors
- [ ] Refactor status.go to use structured errors
- [ ] Refactor restore.go to use structured errors
- [ ] Refactor show.go to use structured errors
- [ ] Refactor worktree_list.go to use structured errors
- [ ] Update cleanup.go if it has similar patterns
- [ ] Update tests to use error type assertions where appropriate
- [ ] Run `make test` to ensure all tests pass
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update CLAUDE.md with error handling guidelines
- [ ] Get developer approval before closing

## Notes

- This ticket emerged from PR #71 review feedback
- Should be done after the test coverage improvements are merged
- Consider whether to make this part of a larger error handling strategy for the entire codebase