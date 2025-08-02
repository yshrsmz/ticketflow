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

## Implementation Details

Successfully implemented a command registration pattern with the following structure:

```go
type Command struct {
    Name         string                                            // Command name for flag set
    MinArgs      int                                               // Minimum number of positional arguments
    MinArgsError string                                            // Optional custom error message for missing args
    SetupFlags   func(*flag.FlagSet) interface{}                  // Optional: Setup command-specific flags
    Validate     func(*flag.FlagSet, interface{}) error          // Optional: Additional validation logic
    Execute      func(context.Context, *flag.FlagSet, interface{}) error // Required: Command execution logic
}
```

The `parseAndExecute` helper function handles:
1. Flag set creation
2. Command-specific flag setup
3. Logging flag addition
4. Flag parsing
5. Logging configuration
6. Argument validation
7. Command execution

## Key Improvements Made

1. **Code Reduction**: Eliminated ~150 lines of repetitive code
2. **Consistency**: All commands now follow the same pattern
3. **Maintainability**: Adding new commands is now straightforward
4. **Documentation**: Added comprehensive field-level documentation
5. **Testing**: Added thorough test coverage for the new functionality
6. **Error Messages**: Custom error messages for better user experience

## Additional Enhancements (Based on Review)

After receiving a positive review from the golang-pro agent, implemented the following improvements:
- Added field-level documentation to the Command struct
- Standardized argument validation using MinArgs instead of custom Validate functions
- Added MinArgsError field for custom error messages
- Created comprehensive tests for parseAndExecute function

## Pull Request

Created PR #30: https://github.com/yshrsmz/ticketflow/pull/30

## Notes

- This is a refactoring task with no functional changes
- Successfully maintained exact same CLI interface
- All existing tests pass without modification
- The refactoring makes it significantly easier to add new commands in the future