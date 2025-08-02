---
priority: 2
description: Refactor repetitive CLI command parsing logic to reduce code duplication
created_at: "2025-08-02T20:45:08+09:00"
started_at: "2025-08-02T20:56:57+09:00"
closed_at: null
related:
    - parent:250801-003207-implement-structured-logging
---

# Ticket Overview

The CLI command parsing logic in `cmd/ticketflow/main.go` is highly repetitive. Each command follows the same pattern:
1. Create flag set
2. Add command-specific flags
3. Add logging flags
4. Parse flags
5. Configure logging
6. Execute command handler

This ticket tracks refactoring this repetitive code into a more maintainable structure.

## Context

This was identified by Copilot during PR #29 review:
> The CLI command parsing logic is highly repetitive. Consider extracting the common pattern of flag parsing and logging configuration into a helper function to reduce code duplication.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Analyze the current command parsing patterns in `runCLI()`
- [x] Design a command registration system or helper functions
- [x] Extract common flag parsing and logging configuration logic
- [x] Refactor each command to use the new pattern
- [x] Ensure backward compatibility with existing CLI interface
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update documentation if necessary
- [ ] Get developer approval before closing

## Proposed Approach

Consider creating a command registry pattern:
```go
type Command struct {
    Name        string
    Handler     func(ctx context.Context, args []string) error
    SetupFlags  func(fs *flag.FlagSet) interface{}
}
```

Or helper functions to reduce duplication:
```go
func parseCommandWithLogging(name string, args []string, setupFlags func(fs *flag.FlagSet) interface{}) (*flag.FlagSet, interface{}, error)
```

## Notes

- This is a refactoring task with no functional changes
- Must maintain exact same CLI interface
- Consider this as a follow-up improvement after structured logging is merged