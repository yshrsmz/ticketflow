---
priority: 3
description: "Migrate remaining commands to new Command interface"
created_at: "2025-08-12T15:29:27+09:00"
started_at: null
closed_at: null
related:
    - parent:250810-003001-refactor-command-interface
    - "blocks:250812-152824-migrate-help-command"
    - "blocks:250812-152902-migrate-init-command"
---

# Migrate remaining commands to new Command interface

Complete the migration of all remaining commands to the new Command interface and finalize the refactoring.

## Commands to Migrate

### Read-Only Commands
- [x] **status** - Show current ticket status (ticket: 250812-231616-migrate-status-command)
- [x] **list** - List tickets with filters (ticket: 250812-213613-migrate-list-command)
- [x] **show** - Display ticket details (ticket: 250813-152930-migrate-show-command)

### State-Changing Commands
- [x] **new** - Create new ticket (with parent flag handling) (ticket: 250813-175042-migrate-new-command) ✅ DONE
- [x] **start** - Start working on ticket (with worktree creation) (ticket: 250813-192015-migrate-start-command) ✅ PR #62 merged
- [ ] **close** - Close current/specified ticket (with reason handling) (ticket: 250814-013846-migrate-close-command) 📝 PR #63 - Refined & ready for implementation
- [ ] **restore** - Restore closed ticket (ticket: 250814-111507-migrate-restore-command) 📋 Created - Next priority

### Complex Commands
- [ ] **worktree** - Manage git worktrees (has subcommands)
- [ ] **cleanup** - Clean up worktrees and branches
- [ ] **migrate** - Migrate ticket dates

## Final Cleanup Tasks

- [ ] Review all new command code and check if there's any implementation/design inconsistency. Report if any
- [ ] Remove old Command struct from command.go
- [ ] Remove parseAndExecute function  
- [ ] Remove entire switch statement from main.go
- [ ] Update all references in documentation
- [ ] Ensure all commands work through registry
- [ ] Run full test suite
- [ ] Update README with new architecture

## Implementation Notes

### Order of Migration (Recommended)
1. Read-only commands first (status, list, show)
2. Simple state-changing commands (new, restore)
3. Complex state-changing commands (start, close)
4. Commands with subcommands (worktree)
5. Utility commands (cleanup, migrate)

### Special Considerations

**For commands with App dependency:**
- Most commands need `cli.App` instance
- Consider dependency injection pattern
- May need factory function for App creation

**For worktree command:**
- Has subcommands (list, clean, etc.)
- May need special handling for subcommand routing

**For commands with complex flags:**
- `new` has parent flag with short form
- `close` has force flag with short form
- Ensure all flag variations work

## Success Criteria

- All commands work exactly as before
- No regression in functionality
- All tests pass
- Clean separation of concerns
- Each command in its own file with tests
- Documentation fully updated

## Progress Summary (2025-08-14)

### Completed Commands
- ✅ **version**, **help**, **init** - Foundation commands
- ✅ **status**, **list**, **show** - Read-only commands  
- ✅ **new** - First state-changing command with parent flag
- ✅ **start** - Complex state-changing with worktree creation

### In Progress
- 🔄 **close** - Ticket refined, PR #63 created, ready for implementation
  - Establishes dual-mode pattern (0 or 1 args)
  - Complex JSON output requirements discovered
  - Time estimate revised to 4-6 hours

### Next Priority
- 📋 **restore** - Ticket created (250814-111507), simplest remaining command
  - 2-3 hours estimated
  - Completes core lifecycle
  - Zero-argument pattern

### Remaining Work
- **3 simple commands**: restore, migrate, cleanup
- **1 complex command**: worktree (has subcommands)
- **Final cleanup**: Remove old code, update docs

### Key Insights from Migration
1. **App methods only return errors** - Commands must gather data for JSON output
2. **Format constants need consolidation** - Currently scattered across commands
3. **Dual-mode complexity** - Optional args add testing complexity
4. **JSON output pattern** - Post-operation data gathering required

## References

- **Migration Guide**: `docs/COMMAND_MIGRATION_GUIDE.md` - Complete step-by-step instructions
- **Example Implementation**: `internal/cli/commands/version.go` - First migrated command
- **Command Executor**: `cmd/ticketflow/executor.go` - Handles new command execution
- **Migration Examples**: `internal/command/migration_example.go` - Pattern examples
- **Interface Definition**: `internal/command/interface.go` - Command interface to implement
- **Registry**: `internal/command/registry.go` - Command registration system
