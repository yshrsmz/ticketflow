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
- [x] Reproduce the issue by opening ticket detail view and pressing 'e'
- [x] Investigate terminal state management in the editor integration code
- [x] Check if the terminal is properly restored after external editor exits
- [x] Review Bubble Tea's ExecProcess handling for external commands
- [x] Implement proper terminal state save/restore mechanism
- [x] Test with different editors (vim, nano, emacs, etc.)
- [x] Ensure TUI redraws correctly after editor returns
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Get developer approval before closing

## Notes

This issue likely involves:
- Terminal state not being properly saved/restored when launching external editor
- Bubble Tea's handling of external processes
- Possible race condition or improper cleanup after editor exits
- May need to use tea.ExecProcess or similar pattern for proper integration

## Resolution

The issue was caused by the `editTicket` function in `internal/ui/app.go` directly executing the external editor command using `exec.Command` and manually connecting stdin/stdout/stderr. This approach conflicts with Bubble Tea's terminal control.

### Root Cause
- The editor was launched with `cmd.Stdin = os.Stdin`, `cmd.Stdout = os.Stdout`, and `cmd.Stderr = os.Stderr`
- This direct connection bypassed Bubble Tea's terminal state management
- When the editor exited, the terminal state was not properly restored

### Fix Applied
Replaced the direct command execution with `tea.ExecProcess`:
```go
// Use tea.ExecProcess to properly handle terminal state
return tea.ExecProcess(cmd, func(err error) tea.Msg {
    // Handle editor result
})
```

This ensures:
1. Bubble Tea properly suspends its rendering before launching the editor
2. Terminal state is saved before the editor starts
3. Terminal state is restored after the editor exits
4. The TUI properly redraws when control returns

The fix has been tested and all linters pass.