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
- [ ] Delete `internal/cli/migrate_dates.go` file entirely (Note: corrected from migrate.go)

### Clean up main.go
- [ ] Remove `migrateFlags` struct definition (lines 170-172)
- [ ] Remove migrate case from switch statement (lines 228-240)
- [ ] Remove `handleMigrateDates` function (lines 336-343)

### Update Help Command (internal/cli/commands/help.go)
- [ ] Remove line 103 from unmigrated commands list
- [ ] Remove lines 160-162 (migrate options section)
- [ ] Remove "migrate" from line 212 in showCommandHelp switch statement

### Documentation Updates
- [ ] Update `docs/COMMAND_MIGRATION_GUIDE.md` line 309 to note migrate was removed instead of migrated
- [ ] Update parent ticket (250812-152927-migrate-remaining-commands) to mark this task as complete

### Verification & Testing
- [ ] Verify all existing tickets have already been migrated (check for any RFC3339Nano dates)
- [ ] Run `make test` to ensure nothing breaks
- [ ] Run `make fmt`, `make vet`, and `make lint`
- [ ] Verify `ticketflow help` no longer shows migrate command
- [ ] Verify `ticketflow migrate` returns "unknown command" error

### Final Steps
- [ ] Update the ticket with insights from resolving this ticket
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