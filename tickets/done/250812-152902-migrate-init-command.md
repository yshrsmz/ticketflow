---
priority: 2
description: Migrate init command to new Command interface
created_at: "2025-08-12T15:29:02+09:00"
started_at: "2025-08-12T18:21:49+09:00"
closed_at: "2025-08-12T20:34:13+09:00"
related:
    - parent:250810-003001-refactor-command-interface
---

# Migrate init command to new Command interface

Migrate the `init` command to use the new Command interface. This command initializes a new ticketflow project.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Create `internal/cli/commands/init.go` implementing the Command interface
- [x] Handle the special case that init doesn't require existing config
- [x] Add unit tests for init command
- [x] Update main.go to use registry for init command
- [x] Remove init case from switch statement
- [x] Test init command in new directory
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

- Init is special: it doesn't require an existing .ticketflow.yaml
- Currently calls `cli.InitCommand(ctx)` directly
- Need to handle this special case in the command implementation
- Follow the version command pattern for the basic structure

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for detailed migration instructions
- Review `internal/cli/commands/version.go` for example implementation
- Check `cmd/ticketflow/executor.go` for command execution pattern
- Note: Init command has special handling in the migration guide

## Implementation Insights

### What Went Well
1. **Simple Migration**: The init command was straightforward to migrate since it has no flags and minimal logic
2. **Code Reuse**: Successfully delegated to existing `cli.InitCommand(ctx)` function, avoiding code duplication
3. **Test Coverage**: Comprehensive unit tests covering all scenarios (new init, already initialized, no git repo)
4. **Clean Removal**: Removing the old switch case and `handleInit` function was clean with no complications

### Key Learnings
1. **Test .gitignore Content**: Initial test failed because it checked for wrong .gitignore content - the actual implementation adds "current-ticket.md" and ".worktrees/", not ".ticketflow.state"
2. **Working Directory Context**: Integration tests need careful handling of working directory changes to avoid affecting the main repository
3. **Linter Formatting**: `go fmt` automatically adds newlines at end of files - this is expected behavior

### Migration Pattern Established
- Commands without flags can have `SetupFlags` return nil
- Commands without validation requirements can have `Validate` return nil immediately
- Special commands that don't require config can still use the same interface pattern

### Next Steps Recommendation
Based on this migration, the `status` command would be a good next candidate as it's also relatively simple but introduces the App dependency pattern that many other commands will need.