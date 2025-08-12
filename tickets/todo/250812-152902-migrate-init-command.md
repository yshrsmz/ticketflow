---
priority: 2
description: "Migrate init command to new Command interface"
created_at: "2025-08-12T15:29:02+09:00"
started_at: null
closed_at: null
related:
    - parent:250810-003001-refactor-command-interface
---

# Migrate init command to new Command interface

Migrate the `init` command to use the new Command interface. This command initializes a new ticketflow project.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Create `internal/cli/commands/init.go` implementing the Command interface
- [ ] Handle the special case that init doesn't require existing config
- [ ] Add unit tests for init command
- [ ] Update main.go to use registry for init command
- [ ] Remove init case from switch statement
- [ ] Test init command in new directory
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update migration guide with completion status
- [ ] Get developer approval before closing

## Implementation Notes

- Init is special: it doesn't require an existing .ticketflow.yaml
- Currently calls `cli.InitCommand(ctx)` directly
- Need to handle this special case in the command implementation
- Follow the version command pattern for the basic structure