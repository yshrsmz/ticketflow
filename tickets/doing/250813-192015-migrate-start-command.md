---
priority: 2
description: Migrate start command to new Command interface
created_at: "2025-08-13T19:20:15+09:00"
started_at: "2025-08-14T00:37:11+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate start command to new Command interface

Migrate the `start` command to use the new Command interface, building on patterns established by previous migrations. This command starts work on a ticket by creating/switching to worktrees and is the second state-changing command to be migrated.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Create `internal/cli/commands/start.go` implementing the Command interface
- [x] Implement App dependency using `cli.NewApp(ctx)` pattern
- [x] Handle positional argument for ticket ID (MinArgs: 1, MaxArgs: 1)
- [x] Implement flags:
  - [x] `--force` / `-f` for force recreate worktree (both long and short forms)
  - [x] `--format` / `-o` for output format (json/text) for consistency
- [x] Add ticket validation in Validate method (exists, not already done)
- [x] Keep Execute method thin, delegating to App.StartTicket
- [x] Add comprehensive unit tests with mock App
- [x] Update main.go to register start command
- [x] Remove start case from switch statement (lines 187-201)
- [x] Test start command functionality with various scenarios:
  - [x] Valid ticket start with worktree creation
  - [x] Valid ticket start without worktree (disabled mode)
  - [x] Ticket not found error (consistent error message format)
  - [x] Ticket already done error
  - [x] Ticket already in doing status
  - [x] Force recreate worktree
  - [x] Parent branch detection for sub-tickets
  - [x] Init commands execution
  - [x] JSON output format
  - [x] Uncommitted changes in current directory
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation
- Located in switch statement around lines 187-201 in main.go
- Calls `handleStart(ctx, ticketID, force)` which delegates to `App.StartTicket`
- Takes one required argument: ticket ID
- Has one flag: `--force` for recreating existing worktrees
- Currently lacks JSON output support (unlike some other commands)

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly to leverage existing `App.StartTicket` method
2. **Positional Arguments**: Required ticket ID argument with validation (MinArgs: 1, MaxArgs: 1)
3. **Ticket Validation**: Ensure ticket exists and is not already done (in Validate method)
4. **Worktree Management**: Handle creation, existence checks, and force recreation
5. **Git Operations**: Branch creation/switching, worktree setup
6. **Output Formatting**: 
   - Consistent success/error messages in text format
   - JSON output support for AI/tooling integration
   - Format flag (`--format`/`-o`) with json/text options
7. **Error Handling**: Clear, consistent error messages matching other commands

### Expected Behavior
- Validates ticket exists and is not done
- Moves ticket from todo → doing status  
- Sets started_at timestamp
- Creates git worktree (if enabled in config)
- Or switches to branch (if worktree disabled)
- Executes init commands if configured
- Handles parent branch detection for sub-tickets
- Supports force recreation of existing worktrees
- Returns structured JSON output when `--format json` is specified:
  ```json
  {
    "ticket_id": "250813-192015-migrate-start-command",
    "status": "doing",
    "worktree_path": "/path/to/worktree",
    "branch": "250813-192015-migrate-start-command",
    "parent_branch": "main",
    "init_commands_executed": true
  }
  ```

## Pattern Building on Previous Migrations

This migration builds on:
1. **State Modification** (from `new`): Changes ticket state and filesystem
2. **Positional Arguments** (from `show`): Single required ticket ID with MinArgs/MaxArgs validation
3. **Boolean Flags**: --force flag with both long and short forms
4. **App Method Reuse**: Leverages existing `App.StartTicket`
5. **JSON Output** (from other commands): Consistent format flag pattern
6. **New Pattern**: Introduces worktree management patterns

## New Patterns to Establish

1. **Worktree Operations**: First command to create/manage git worktrees
2. **Git State Validation**: Checking for uncommitted changes
3. **Conditional Execution**: Different paths for worktree vs non-worktree modes
4. **Init Commands**: Running configured initialization commands
5. **Force Operations**: Pattern for force flags that override safety checks

## Estimated Time
**3-4 hours** based on:
- `new` command took ~2 hours (state-changing)
- `start` adds worktree management complexity
- Need to handle both worktree and non-worktree modes
- More complex git operations than previous commands

## Why This Command Next?

1. **Natural Workflow**: Follows `new` in typical user flow
2. **Incremental Complexity**: Builds on state-changing patterns
3. **Strategic Foundation**: Establishes worktree patterns for future commands
4. **High Impact**: Frequently used core command
5. **Paired Commands**: Sets up patterns needed for `close` command

## Technical Considerations

1. **Worktree Validation**: Check existence, handle force recreation
2. **Git Branch Management**: Parent branch detection, creation strategies
3. **State Transitions**: Atomic updates to ticket status
4. **Error Recovery**: Handle partial state changes gracefully
5. **Testing Complexity**: Mock git operations and filesystem
6. **Config Variations**: Test with worktree enabled/disabled
7. **Init Command Execution**: Handle command failures gracefully
8. **JSON Output Structure**: Define consistent schema for success and error cases
9. **Error Message Consistency**: Match format from other migrated commands
10. **Flag Validation**: Ensure format flag accepts only "json" or "text"

## Dependencies
- Builds on patterns from: `new`, `show`, `status` commands
- Will inform patterns for: `close`, `cleanup`, `worktree` commands

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/new.go` for state-changing patterns and JSON output
- Check current `handleStart` implementation in main.go (lines 187-201, 291-298)
- Review `TestApp_StartTicket_WithMocks` for test patterns
- New command PR for reference on state-changing commands