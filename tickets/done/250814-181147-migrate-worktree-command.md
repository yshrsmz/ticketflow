---
priority: 3
description: Migrate worktree command with subcommands to new Command interface
created_at: "2025-08-14T18:11:47+09:00"
started_at: "2025-08-15T10:41:24+09:00"
closed_at: "2025-08-15T15:10:19+09:00"
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate Worktree Command to New Interface

Migrate the worktree command and all its subcommands from the old switch-based system to the new Command interface pattern. This is the most complex remaining migration as it involves multiple subcommands.

## Context

The worktree command is the last remaining command to migrate from the old switch-based system. It has exactly two subcommands:
- `worktree list` - List all worktrees (with JSON output support)
- `worktree clean` - Clean up orphaned worktrees

This migration requires implementing the subcommand pattern within the new Command interface structure. The pattern is already documented in COMMAND_MIGRATION_GUIDE.md (lines 185-201).

## Implementation Details

The worktree command needs to:
1. Act as a parent command that delegates to subcommands
2. Implement each subcommand as its own Command interface
3. Handle subcommand routing and help text
4. Maintain backward compatibility with existing usage patterns

### Architecture Notes (from analysis):
- The actual implementation is straightforward (~100 lines total for both subcommands)
- `ListWorktrees()`: ~30 lines with JSON/text output formatting
- `CleanWorktrees()`: ~50 lines for orphaned worktree cleanup
- The perceived complexity is architectural (parent/child command pattern), not implementation difficulty
- This completes the entire command interface migration project

## Tasks

- [x] Analyze existing worktree command implementation and all subcommands
- [x] Create internal/cli/commands/worktree.go as the parent command
- [x] Implement subcommand routing pattern (following COMMAND_MIGRATION_GUIDE.md lines 185-201)
- [x] Create separate command files for each subcommand:
  - [x] internal/cli/commands/worktree_list.go
  - [x] internal/cli/commands/worktree_clean.go
- [x] Ensure help text works correctly for parent and subcommands
- [x] Update main.go to use the new worktree command structure
- [x] Remove old worktree implementation
- [x] Add/update tests for the new implementation
- [x] Test all subcommands thoroughly
- [x] Run `make test` to ensure all tests pass
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update COMMAND_MIGRATION_GUIDE.md with any new learnings about subcommand pattern
- [x] Update parent ticket (250812-152927-migrate-remaining-commands) to mark this task as complete
- [x] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

While this appears to be the most complex migration due to subcommands, the analysis reveals it's actually quite manageable:
- Total implementation is only ~100 lines of business logic
- The subcommand pattern is already documented in COMMAND_MIGRATION_GUIDE.md
- This is the final piece to complete the entire command migration project

Priority is set to 3 (higher) to ensure timely completion of the migration project.

### Estimated Effort
- Implementation: 2-3 hours
- Testing: 1 hour
- Total: 3-4 hours

## Implementation Insights

### Key Learnings from Implementation

1. **Subcommand Pattern Works Well**: The parent command pattern with subcommand routing is clean and maintainable. Each subcommand is a full Command implementation, making them testable in isolation.

2. **Flag Handling for Subcommands**: The parent command needs to create a new FlagSet for each subcommand and handle parsing separately. This was implemented successfully in the Execute method.

3. **Validation Edge Cases**: The list subcommand's Validate method needed careful handling of nil flags and empty format strings. Default values should be applied in Execute rather than Validate.

4. **Test Structure**: Each command (parent and subcommands) needs its own test file. The parent command tests focus on routing, while subcommand tests focus on their specific logic.

5. **Documentation Update**: The COMMAND_MIGRATION_GUIDE.md was enhanced with detailed subcommand implementation patterns including flag parsing and validation flow.

### Migration Completion

**This completes the entire command interface migration project!** All 12 commands have been successfully migrated to the new Command interface pattern:
- Foundation: version, help, init
- Read-only: status, list, show
- State-changing: new, start, close, restore, cleanup
- With subcommands: worktree (list, clean)

The old switch statement in main.go has been completely eliminated, and all commands now use the registry pattern.