---
priority: 2
description: Migrate help command to new Command interface
created_at: "2025-08-12T15:28:24+09:00"
started_at: "2025-08-12T16:59:50+09:00"
closed_at: null
related:
    - parent:250810-003001-refactor-command-interface
---

# Migrate help command to new Command interface

Migrate the `help` command (including `-h` and `--help` aliases) to use the new Command interface following the pattern established with the version command migration.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Create `internal/cli/commands/help.go` implementing the Command interface
- [x] Implement help text generation from registered commands
- [x] Add unit tests for help command
- [x] Update main.go to handle help aliases through registry
- [x] Remove help case from switch statement
- [x] Test all help command variations (help, -h, --help)
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

- Follow the same pattern as version command
- Help command should list all available commands from registry
- Consider adding command descriptions to the help output
- Ensure backward compatibility during migration

## References

- See `docs/COMMAND_MIGRATION_GUIDE.md` for detailed migration instructions
- Review `internal/cli/commands/version.go` for example implementation
- Check `cmd/ticketflow/executor.go` for command execution pattern