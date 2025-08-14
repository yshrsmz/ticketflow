---
priority: 2
description: Remove deprecated migrate command from codebase
created_at: "2025-08-14T18:10:27+09:00"
started_at: "2025-08-14T18:19:39+09:00"
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Remove Migrate Command

The migrate command is no longer needed as all migrations have been completed. This command was used for one-time date format migrations in ticket files, but there should be no more migrations required.

## Context

The migrate command was originally created to handle migrations of ticket file formats (e.g., date format changes). Since the codebase has stabilized and no further migrations are planned, this command adds unnecessary complexity and should be removed.

## Tasks

### Core Implementation Removal
- [x] Delete `internal/cli/migrate_dates.go` file entirely (Note: corrected from migrate.go)

### Clean up main.go
- [x] Remove `migrateFlags` struct definition (lines 170-172)
- [x] Remove migrate case from switch statement (lines 228-240)
- [x] Remove `handleMigrateDates` function (lines 336-343)

### Update Help Command (internal/cli/commands/help.go)
- [x] Remove line 103 from unmigrated commands list
- [x] Remove lines 160-162 (migrate options section)
- [x] Remove "migrate" from line 212 in showCommandHelp switch statement

### Documentation Updates
- [x] Update `docs/COMMAND_MIGRATION_GUIDE.md` line 309 to note migrate was removed instead of migrated
- [x] Update parent ticket (250812-152927-migrate-remaining-commands) to mark this task as complete
- [x] Update CLAUDE.md to clarify ticketflow binary location

### Verification & Testing
- [x] Verify all existing tickets have already been migrated (check for any RFC3339Nano dates)
- [x] Run `make test` to ensure nothing breaks
- [x] Run `make fmt`, `make vet`, and `make lint`
- [x] Verify `ticketflow help` no longer shows migrate command
- [x] Verify `ticketflow migrate` returns "unknown command" error

### Final Steps
- [x] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

This is part of the command interface refactoring effort. Instead of migrating the migrate command to the new interface, we're removing it entirely as it's no longer needed.

## Analysis Insights

Based on codebase analysis:
- The actual file is `internal/cli/migrate_dates.go` (not migrate.go as originally stated)
- Additional cleanup needed in main.go: `migrateFlags` struct and `handleMigrateDates` function
- No dedicated test files exist for the migrate command (simplifies removal)
- Help command has 3 references that need removal
- Estimated effort: 1-2 hours including testing and documentation

## Implementation Summary

Successfully removed the migrate command from the codebase:

1. **Verified Migration Status**: Confirmed all existing tickets have been migrated to RFC3339 format (no nanoseconds)
2. **Code Removal**: 
   - Deleted `internal/cli/migrate_dates.go` (107 lines)
   - Removed `migrateFlags` struct, migrate case, and `handleMigrateDates` function from main.go
   - Cleaned up all help command references
3. **Documentation Updates**:
   - Updated COMMAND_MIGRATION_GUIDE.md to note removal instead of migration
   - Updated parent ticket to track completion
   - Added clarification to CLAUDE.md about binary location
4. **Testing & Verification**:
   - All tests pass (`make test`)
   - Code formatted and linted (`make fmt`, `make vet`, `make lint`)
   - Verified `ticketflow migrate` returns "unknown command" error
   - Confirmed migrate command no longer appears in help output

**Total commits**: 5 focused commits tracking each major change
**Actual effort**: ~45 minutes (under the 1-2 hour estimate)