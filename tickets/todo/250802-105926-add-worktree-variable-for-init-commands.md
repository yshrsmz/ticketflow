---
priority: 2
description: "Add worktree path variable support in init_commands configuration"
created_at: "2025-08-02T10:59:26+09:00"
started_at: null
closed_at: null
---

# Ticket Overview

Add support for a worktree path variable that can be used in the `init_commands` configuration. This will allow developers to automatically open their editor in the correct worktree directory when starting work on a ticket.

## Tasks
- [ ] Add worktree path variable (e.g., `{{worktree}}` or `$WORKTREE`) to be available in init_commands
- [ ] Update the command execution logic to replace the variable with the actual worktree path
- [ ] Test with editor commands like `code {{worktree}}` or `vim {{worktree}}`
- [ ] Update configuration documentation to explain the new variable
- [ ] Add example to .ticketflow.yaml showing editor launch usage
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update CLAUDE.md if necessary
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

This feature will enable configurations like:
```yaml
worktree:
  init_commands:
    - code {{worktree}}  # Opens VS Code in the worktree directory
    - cd {{worktree}} && npm install  # Run setup in worktree
```

The variable should be available for all commands in the init_commands list.