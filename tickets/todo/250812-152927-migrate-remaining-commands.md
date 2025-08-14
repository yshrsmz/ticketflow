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
- [x] **close** - Close current/specified ticket (with reason handling) (ticket: 250814-013846-migrate-close-command) ✅ Implementation complete, pending final verification
- [x] **restore** - Restore closed ticket (ticket: 250814-111507-migrate-restore-command) ✅ COMPLETED (2025-08-14)

### Complex Commands
- [ ] **worktree** - Manage git worktrees (has subcommands) (ticket: 250814-181147-migrate-worktree-command) 📋 Created
- [x] **cleanup** - Clean up worktrees and branches (ticket: 250814-181107-migrate-cleanup-command) ✅ COMPLETED (2025-08-15)
- [x] **migrate** - ~~TO BE REMOVED - No longer needed~~ ✅ REMOVED (ticket: 250814-181027-remove-migrate-command)

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

## Progress Summary

### Completed Commands (10 total)
- ✅ **version**, **help**, **init** - Foundation commands
- ✅ **status**, **list**, **show** - Read-only commands  
- ✅ **new** - First state-changing command with parent flag
- ✅ **start** - Complex state-changing with worktree creation
- ✅ **close** - Dual-mode pattern (0 or 1 args)
- ✅ **restore** - Zero-argument pattern

### Recently Completed (2025-08-14)
- ✅ **App Method Return Values** - Refactor App methods to return primary entities (ticket: 250814-121422)
  - **COMPLETED** - All App methods now return entities
  - Eliminated re-fetching for JSON output (50% I/O reduction)
  - Updated new, start, close commands to use returned entities
  - Added helper methods for derived data
  - Comprehensive tests and documentation updated
- ✅ **restore** - Command migration completed (ticket: 250814-111507)
  - Implemented with entity return pattern from the start
  - Zero-argument pattern, simple implementation
  - All tests passing

### Current Status (2025-08-14 19:00)
**Migration Progress: ~90% Complete**
- **10 commands migrated**: version, help, init, status, list, show, new, start, close, restore
- **1 command removed**: migrate (no longer needed - all tickets already migrated)
- **2 commands remaining**: 
  - Migrate cleanup command
  - Migrate worktree command (with subcommands)

### Next Priority
- ✅ **remove-migrate-command** (250814-181027) - COMPLETED: Removed all migrate command code
- 📋 **migrate-cleanup-command** (250814-181107) - Straightforward migration
- 📋 **migrate-worktree-command** (250814-181147) - Complex with subcommands, highest priority (3)

### Remaining Work
- **2 commands to migrate**: cleanup, worktree (with subcommands)
- **Final cleanup**: Remove old code, switch statement, update docs

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
