---
priority: 2
description: Migrate close command to new Command interface
created_at: "2025-08-14T01:38:46+09:00"
started_at: "2025-08-14T09:59:44+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate close command to new Command interface

Migrate the `close` command to use the new Command interface, completing the core ticket lifecycle commands (new → start → close). This command closes tickets and handles the final state transition.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Create `internal/cli/commands/close.go` implementing the Command interface
- [ ] Implement App dependency using `cli.NewApp(ctx)` pattern
- [ ] Handle optional positional argument for ticket ID (0 or 1 args)
- [ ] Implement flags:
  - [ ] `--force` / `-f` to force close even with uncommitted changes
  - [ ] `--reason` for closing reason (added to commit message)
  - [ ] `--format` / `-o` for output format (json/text)
- [ ] Add validation in Validate method:
  - [ ] Check ticket exists if ID provided
  - [ ] Verify ticket is not already closed
  - [ ] Validate uncommitted changes (unless forced)
- [ ] Keep Execute method thin, delegating to App.CloseTicket
- [ ] Handle dual behavior:
  - [ ] No args: close current ticket in worktree
  - [ ] With ID: close specific ticket
- [ ] Add comprehensive unit tests with mock App
- [ ] Update main.go to register close command
- [ ] Remove close case from switch statement (lines 202-219)
- [ ] Test close command functionality with various scenarios:
  - [ ] Close current ticket in worktree
  - [ ] Close specific ticket by ID
  - [ ] Error when ticket not found
  - [ ] Error when ticket already closed
  - [ ] Force close with uncommitted changes
  - [ ] Close with reason provided
  - [ ] JSON output format
  - [ ] No current ticket error (when no args in main repo)
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation
- Located in switch statement around lines 202-219 in main.go
- Calls `handleClose(ctx, ticketID, force, reason)` which delegates to `App.CloseTicket`
- Takes optional argument: ticket ID (if not provided, uses current ticket)
- Has flags:
  - `--force` / `-f` for forcing close with uncommitted changes
  - `--reason` for adding close reason to commit message
- Currently lacks JSON output support

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly to leverage existing `App.CloseTicket` method
2. **Optional Positional Arguments**: Handle 0 or 1 arguments (MinArgs: 0, MaxArgs: 1)
3. **Dual Behavior**:
   - No args: Detect and close current ticket from worktree
   - With ID: Close specified ticket
4. **Validation**:
   - Ticket exists and is not already closed
   - Check for uncommitted changes (unless forced)
   - Current ticket detection when no ID provided
5. **Output Formatting**:
   - Consistent success/error messages in text format
   - JSON output support for AI/tooling integration
   - Format flag (`--format`/`-o`) with json/text options
6. **Commit Integration**: Handle reason flag for commit message enhancement
7. **Error Handling**: Clear, consistent error messages matching other commands

### Expected Behavior
- Validates ticket exists and is not already closed
- Moves ticket from doing → done status
- Sets closed_at timestamp
- Creates "Close ticket" commit (with optional reason)
- Handles uncommitted changes check (bypassable with --force)
- Returns structured JSON output when `--format json` is specified:
  ```json
  {
    "ticket_id": "250814-013846-migrate-close-command",
    "status": "done",
    "closed_at": "2025-08-14T02:00:00+09:00",
    "close_reason": "Migration completed successfully",
    "commit_created": true,
    "force_used": false
  }
  ```

## Pattern Building on Previous Migrations

This migration builds on:
1. **State Modification** (from `new`, `start`): Changes ticket state and filesystem
2. **Optional Positional Arguments**: New pattern - 0 or 1 arguments
3. **Boolean Flags** (from `start`): --force flag with both long and short forms
4. **String Flags**: --reason flag for additional context
5. **App Method Reuse**: Leverages existing `App.CloseTicket`
6. **JSON Output** (from other commands): Consistent format flag pattern
7. **Current Context Detection**: Determining current ticket from worktree

## New Patterns to Establish

1. **Optional Positional Arguments**: First command with 0-1 positional args
2. **Current Context Detection**: Inferring ticket from current directory
3. **Commit Message Enhancement**: Adding custom reason to commits
4. **Dual Mode Operation**: Different behavior based on argument presence
5. **Uncommitted Changes Validation**: Safety check with force override

## Estimated Time
**3-4 hours** based on:
- Similar complexity to `start` command
- Additional complexity from dual behavior modes
- Current ticket detection logic
- Uncommitted changes validation

## Why This Command Next?

1. **Completes Core Lifecycle**: new → start → close forms the essential workflow
2. **Natural Progression**: Logical next step after `start` command
3. **Moderate Complexity**: Right level of challenge after `start`
4. **High User Value**: Frequently used command in daily workflow
5. **Unblocks Related Commands**: `restore` naturally follows `close`
6. **Pattern Completion**: Establishes remaining state-change patterns

## Technical Considerations

1. **Current Ticket Detection**: Determine ticket from worktree directory
2. **Git Status Check**: Validate clean working directory
3. **Force Flag Handling**: Override safety checks when forced
4. **Reason Integration**: Append reason to commit message properly
5. **State Transitions**: Atomic updates to ticket status
6. **Dual Mode Logic**: Handle both current and specific ticket paths
7. **Error Recovery**: Handle partial state changes gracefully
8. **Testing Complexity**: Mock both modes of operation
9. **JSON Output Structure**: Consistent schema for both modes
10. **Error Message Consistency**: Match format from other migrated commands

## Dependencies
- Builds on patterns from: `new`, `start`, `show` commands
- Will inform patterns for: `restore`, `cleanup` commands
- Related to worktree management established in `start`

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/start.go` for similar state-changing patterns
- Check current `handleClose` implementation in main.go (lines 202-219, 299-306)
- Review `TestApp_CloseTicket_WithMocks` for test patterns
- Start command PR #62 for reference on state-changing commands with force flag