---
priority: 2
description: TUI breaks when returning from external editor in ticket detail view
created_at: "2025-07-28T17:17:59+09:00"
started_at: "2025-07-28T17:21:04+09:00"
closed_at: null
---

# Ticket Overview

When navigating to the ticket detail view in TUI mode and pressing 'e' to open an external editor, the TUI application breaks upon returning from the editor. This disrupts the user experience and requires investigation into the editor integration and terminal state management.

## Tasks
- [ ] Reproduce the issue by opening ticket detail view and pressing 'e'
- [ ] Investigate terminal state management in the editor integration code
- [ ] Check if the terminal is properly restored after external editor exits
- [ ] Review Bubble Tea's ExecProcess handling for external commands
- [ ] Implement proper terminal state save/restore mechanism
- [ ] Test with different editors (vim, nano, emacs, etc.)
- [ ] Ensure TUI redraws correctly after editor returns
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Get developer approval before closing

## Notes

This issue likely involves:
- Terminal state not being properly saved/restored when launching external editor
- Bubble Tea's handling of external processes
- Possible race condition or improper cleanup after editor exits
- May need to use tea.ExecProcess or similar pattern for proper integration