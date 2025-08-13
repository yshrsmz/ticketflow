---
priority: 2
description: Migrate start command to new Command interface
created_at: "2025-08-13T19:20:15+09:00"
started_at: null
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate start command to new Command interface

Migrate the `start` command to use the new Command interface, building on patterns established by previous migrations. This command starts work on a ticket by creating/switching to worktrees and is the second state-changing command to be migrated.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Create `internal/cli/commands/start.go` implementing the Command interface
- [ ] Implement App dependency using `cli.NewApp(ctx)` pattern
- [ ] Handle positional argument for ticket ID (MinArgs: 1)
- [ ] Implement flags:
  - [ ] `--force` / `-f` for force recreate worktree
- [ ] Add ticket validation (exists, not already done)
- [ ] Handle worktree creation and management
- [ ] Handle non-worktree mode (branch switching)
- [ ] Add comprehensive unit tests with mock App
- [ ] Update main.go to register start command
- [ ] Remove start case from switch statement
- [ ] Test start command functionality with various scenarios:
  - [ ] Valid ticket start with worktree creation
  - [ ] Valid ticket start without worktree (disabled mode)
  - [ ] Ticket not found error
  - [ ] Ticket already done error
  - [ ] Force recreate worktree
  - [ ] Parent branch detection for sub-tickets
  - [ ] Init commands execution
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

### Current Implementation
- Located in switch statement around line 210-230 in main.go
- Calls `handleStart(ctx, ticketID, force)`
- Takes one required argument: ticket ID
- Has one flag: `--force` for recreating existing worktrees

### Migration Requirements
1. **App Dependency**: Use `cli.NewApp(ctx)` directly to leverage existing `App.StartTicket` method
2. **Positional Arguments**: Required ticket ID argument with validation
3. **Ticket Validation**: Ensure ticket exists and is not already done
4. **Worktree Management**: Handle creation, existence checks, and force recreation
5. **Git Operations**: Branch creation/switching, worktree setup
6. **Output Formatting**: Consistent success/error messages
7. **Error Handling**: Clear messages for various failure scenarios

### Expected Behavior
- Validates ticket exists and is not done
- Moves ticket from todo → doing status
- Sets started_at timestamp
- Creates git worktree (if enabled in config)
- Or switches to branch (if worktree disabled)
- Executes init commands if configured
- Handles parent branch detection for sub-tickets
- Supports force recreation of existing worktrees

## Pattern Building on Previous Migrations

This migration builds on:
1. **State Modification** (from `new`): Changes ticket state and filesystem
2. **Positional Arguments** (from `show`): Single required ticket ID
3. **Boolean Flags**: Simpler than `new`, just --force flag
4. **App Method Reuse**: Leverages existing `App.StartTicket`
5. **New Pattern**: Introduces worktree management patterns

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

## Dependencies
- Builds on patterns from: `new`, `show`, `status` commands
- Will inform patterns for: `close`, `cleanup`, `worktree` commands

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for migration patterns
- Review `internal/cli/commands/new.go` for state-changing patterns
- Check current `handleStart` implementation in main.go
- New command PR for reference on state-changing commands