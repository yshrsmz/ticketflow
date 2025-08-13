---
priority: 2
description: Migrate show command to new Command interface
created_at: "2025-08-13T15:29:30+09:00"
started_at: "2025-08-13T15:49:00+09:00"
closed_at: "2025-08-13T17:56:43+09:00"
related:
    - parent:250812-213613-migrate-list-command
---

# Migrate show command to new Command interface

Migrate the `show` command to use the new Command interface, continuing the pattern established by the status and list command migrations. This command displays detailed information about a specific ticket.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Create `internal/cli/commands/show.go` implementing the Command interface
- [x] Move show logic from handleShow into ShowCommand.Execute (inline, not as App method)
- [x] Implement MinArgs validation for positional ticket ID argument
- [x] Implement `--format` flag for output format (text/json)
- [x] Preserve exact JSON output structure for compatibility
- [x] Handle time formatting with nil checks
- [x] Add ticket ID validation and error handling for missing tickets
- [x] Add comprehensive unit tests with mock Manager
- [x] Update main.go to register show command
- [x] Remove show case from switch statement
- [x] Test show command functionality with various scenarios:
  - [x] Valid full ticket ID
  - [x] Valid partial ticket ID
  - [x] Ambiguous partial ID (multiple matches)
  - [x] Non-existent ticket ID
  - [x] Both output formats (text/json)
  - [x] Nil timestamp handling
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update migration guide with completion status
- [x] Address code review feedback:
  - [x] Add defensive type assertion checks
  - [x] Handle empty format string for backward compatibility
  - [x] Validate no extra arguments after ticket ID
  - [x] Improve test coverage for edge cases
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation
- Located in switch statement around line 215-230 in main.go
- Calls `handleShow(ctx, ticketID, format)`
- Takes one required argument: ticket ID (can be partial)
- Has one flag: `--format` for output format (text/json)

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly (following established pattern)
2. **Positional Arguments**: First command to require a positional argument (use MinArgs validation)
3. **Ticket Resolution**: Handle partial ticket ID matching via Manager.Get()
4. **Error Handling**: Proper error messages for missing or ambiguous tickets
5. **Output Formatting**: Support both text and JSON output formats - preserve exact structure
6. **Implementation**: Move logic directly from handleShow, not via App.ShowTicket method

### Expected Behavior
- Shows detailed ticket information including:
  - ID, status, priority
  - Description
  - Created/started/closed timestamps
  - Related tickets (parent, blocks, etc.)
  - Full ticket content
- Supports partial ticket ID matching (e.g., "250813" matches "250813-152930-migrate-show-command")
- Returns clear error for ambiguous partial matches
- Returns helpful error for non-existent tickets

## Pattern Differences from Previous Migrations

This is the first migrated command that:
1. **Requires positional arguments** - Use MinArgs validation
2. **Operates on a specific resource** - Need to fetch and display a single ticket
3. **Has complex output** - Full ticket details vs. simple lists

## Estimated Time
**45-60 minutes** based on:
- List command took 45 minutes (actual)
- First command with positional arguments (new pattern to establish)
- Complex output formatting with time handling
- Need to preserve exact JSON structure for compatibility
- Comprehensive test coverage for various scenarios

## Why This Command Next?

1. **Logical Flow**: Natural progression from `list` (all tickets) to `show` (one ticket)
2. **New Patterns**: Introduces positional argument handling for future commands
3. **Read-Only**: Safe operation with no data modifications
4. **High Usage**: Frequently used command in daily workflow
5. **Foundation**: Establishes patterns for commands that operate on specific tickets

## Technical Considerations

1. **Argument Validation**: Implement MinArgs checking in Validate method
2. **Ticket Resolution**: Use app.Manager.Get() which handles partial matching
3. **Error Messages**: Ensure clear, actionable error messages for common issues
4. **Output Format**: Move formatting logic from handleShow directly into Execute method
5. **Implementation Note**: No App.ShowTicket method exists - inline the logic
6. **JSON Compatibility**: Preserve exact structure for tools consuming the output
7. **Testing**: Mock Manager.Get() for various scenarios (found, not found, ambiguous)

## Implementation Insights

### Key Learnings
1. **Type Safety**: Always use safe type assertions with the `, ok` pattern to prevent panics
2. **Backward Compatibility**: Empty string flags should default to sensible values rather than error
3. **Argument Validation**: Validate both minimum AND maximum arguments to prevent misuse
4. **Test Structure**: Using `interface{}` for flags in tests allows testing type assertion failures
5. **Code Review Value**: The golang-pro review caught important edge cases and improvements

### Actual Time
**~50 minutes** - Aligned well with the 45-60 minute estimate after refinement

### Patterns Established
- First command with positional arguments using MinArgs pattern
- Validation of extra arguments to prevent silent failures
- Defensive programming with type assertions in both Validate and Execute
- Backward compatibility approach for empty/missing flag values

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/list.go` for Command interface pattern
- Check current `handleShow` implementation in main.go (line ~340)
- List command PR: #59 for reference implementation