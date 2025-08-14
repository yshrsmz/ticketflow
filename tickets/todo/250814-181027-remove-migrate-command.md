---
priority: 2
description: "Remove deprecated migrate command from codebase"
created_at: "2025-08-14T18:10:27+09:00"
started_at: null
closed_at: null
related:
    - parent:250812-152927-migrate-remaining-commands
---

# Remove Migrate Command

The migrate command is no longer needed as all migrations have been completed. This command was used for one-time date format migrations in ticket files, but there should be no more migrations required.

## Context

The migrate command was originally created to handle migrations of ticket file formats (e.g., date format changes). Since the codebase has stabilized and no further migrations are planned, this command adds unnecessary complexity and should be removed.

## Tasks

- [ ] Remove migrate command from internal/cli/migrate.go
- [ ] Remove migrate case from main.go switch statement
- [ ] Remove any migrate-related tests
- [ ] Remove migrate from help text and documentation
- [ ] Verify no other code references the migrate command
- [ ] Run `make test` to ensure nothing breaks
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update COMMAND_MIGRATION_GUIDE.md to note migrate was removed instead of migrated
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

This is part of the command interface refactoring effort. Instead of migrating the migrate command to the new interface, we're removing it entirely as it's no longer needed.