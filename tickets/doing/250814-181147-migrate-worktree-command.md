---
priority: 3
description: Migrate worktree command with subcommands to new Command interface
created_at: "2025-08-14T18:11:47+09:00"
started_at: "2025-08-15T10:41:24+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate Worktree Command to New Interface

Migrate the worktree command and all its subcommands from the old switch-based system to the new Command interface pattern. This is the most complex remaining migration as it involves multiple subcommands.

## Context

The worktree command is the most complex of the remaining commands to migrate because it has multiple subcommands:
- `worktree list` - List all worktrees
- `worktree clean` - Clean up orphaned worktrees
- Additional worktree operations may exist

This migration requires implementing the subcommand pattern within the new Command interface structure.

## Implementation Details

The worktree command needs to:
1. Act as a parent command that delegates to subcommands
2. Implement each subcommand as its own Command interface
3. Handle subcommand routing and help text
4. Maintain backward compatibility with existing usage patterns

## Tasks

- [ ] Analyze existing worktree command implementation and all subcommands
- [ ] Create internal/cli/worktree_command.go as the parent command
- [ ] Implement subcommand routing pattern (similar to how git handles subcommands)
- [ ] Create separate command files for each subcommand:
  - [ ] internal/cli/worktree_list_command.go
  - [ ] internal/cli/worktree_clean_command.go
  - [ ] Any other worktree subcommands found
- [ ] Ensure help text works correctly for parent and subcommands
- [ ] Update main.go to use the new worktree command structure
- [ ] Remove old worktree implementation
- [ ] Add/update tests for the new implementation
- [ ] Test all subcommands thoroughly
- [ ] Run `make test` to ensure all tests pass
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update COMMAND_MIGRATION_GUIDE.md with notes about subcommand pattern
- [ ] Update parent ticket (250812-152927-migrate-remaining-commands) to mark this task as complete
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

This is the most complex migration due to subcommands. Consider looking at how other CLI tools handle subcommands in Go (e.g., cobra-based tools) for patterns, though we need to fit within our existing Command interface structure.

Priority is set to 3 (higher) because this is more complex and may uncover issues with the Command interface pattern that need to be addressed.