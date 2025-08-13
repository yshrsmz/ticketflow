---
priority: 2
description: "Migrate show command to new Command interface"
created_at: "2025-08-13T15:29:30+09:00"
started_at: null
closed_at: null
related:
    - parent:250812-213613-migrate-list-command
---

# Migrate show command to new Command interface

Migrate the `show` command to use the new Command interface, continuing the pattern established by the status and list command migrations. This command displays detailed information about a specific ticket.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Create `internal/cli/commands/show.go` implementing the Command interface
- [ ] Implement App dependency using `cli.NewApp(ctx)` pattern
- [ ] Handle positional argument for ticket ID (MinArgs: 1)
- [ ] Implement `--format` flag for output format (text/json)
- [ ] Add ticket ID validation and error handling for missing tickets
- [ ] Add comprehensive unit tests with mock App
- [ ] Update main.go to register show command
- [ ] Remove show case from switch statement
- [ ] Test show command functionality with various scenarios:
  - [ ] Valid ticket ID
  - [ ] Invalid/missing ticket ID
  - [ ] Partial ticket ID matching
  - [ ] Different output formats
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation
- Located in switch statement around line 215-230 in main.go
- Calls `handleShow(ctx, ticketID, format)`
- Takes one required argument: ticket ID (can be partial)
- Has one flag: `--format` for output format (text/json)

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly (following established pattern)
2. **Positional Arguments**: First command to require a positional argument
3. **Ticket Resolution**: Handle partial ticket ID matching via Manager.Get()
4. **Error Handling**: Proper error messages for missing or ambiguous tickets
5. **Output Formatting**: Support both text and JSON output formats

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
**20-25 minutes** based on:
- List command took 45 minutes (actual)
- Show is simpler (fewer flags, no status validation)
- But adds positional argument handling
- Similar test complexity

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
4. **Output Format**: Reuse existing formatting logic from handleShow
5. **Testing**: Mock Manager.Get() for various scenarios (found, not found, ambiguous)

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/list.go` for Command interface pattern
- Check current `handleShow` implementation in main.go (line ~340)
- List command PR: #59 for reference implementation