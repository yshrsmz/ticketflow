---
priority: 2
description: "Migrate cleanup command to new Command interface"
created_at: "2025-08-14T18:11:07+09:00"
started_at: null
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Migrate Cleanup Command to New Interface

Migrate the cleanup command from the old switch-based system to the new Command interface pattern, following the established migration guide.

## Context

The cleanup command removes worktrees and branches for closed tickets. It's one of the last remaining commands that needs to be migrated to the new Command interface as part of the architectural refactoring effort.

## Implementation Details

The cleanup command:
- Takes a ticket ID as argument
- Validates the ticket exists and is closed
- Removes the associated git worktree (if it exists)
- Removes the associated git branch (if it exists)
- Provides feedback on what was cleaned up

## Tasks

- [ ] Create internal/cli/cleanup_command.go implementing the Command interface
- [ ] Move cleanup logic from existing implementation to new Execute method
- [ ] Follow the pattern established in other migrated commands (e.g., close_command.go)
- [ ] Update main.go to use the new cleanup command
- [ ] Remove old cleanup implementation
- [ ] Add/update tests for the new implementation
- [ ] Run `make test` to ensure all tests pass
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update COMMAND_MIGRATION_GUIDE.md to mark cleanup as migrated
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

Reference the COMMAND_MIGRATION_GUIDE.md for the standard migration pattern. The cleanup command is relatively straightforward with no subcommands, similar to the close command that was already migrated.