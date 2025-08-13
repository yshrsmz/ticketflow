---
priority: 2
description: "Add progress indicator for cleanup command"
created_at: "2025-08-11T21:54:24+09:00"
started_at: null
closed_at: null
related:
    - parent:250810-002848-refactor-concurrent-directory-ops
---

# Progress Indicators for Large Operations

Add optional progress indicators for operations that take more than 1 second.

## Background

While most TicketFlow operations are instant (3-30ms), the `cleanup` command can legitimately take several seconds when removing many large git worktrees. Users should have feedback during this operation.

## Tasks

- [ ] Focus specifically on the `cleanup` command where it's actually needed
- [ ] Implement simple progress indicator for worktree removal:
  - Show "Removing worktree X of Y: [worktree-name]"
  - Simple counter, no need for fancy progress bars
- [ ] Ensure progress output goes to stderr (not stdout)
- [ ] Respect quiet mode if implemented
- [ ] Keep it simple - no need for complex terminal detection
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update README.md
- [ ] Get developer approval before closing

## Implementation Notes

- Use existing TUI libraries (Bubble Tea) for consistent styling
- Progress should not interfere with JSON output mode
- Consider using `golang.org/x/term` for terminal detection
- Progress updates should be rate-limited (e.g., max 10 updates/second)

## Example Progress Display

```
Loading tickets... [####------] 40% (400/1000) 2.3s
```

Or with spinner:
```
⠼ Loading 1,234 tickets... (5.2s)
```

## UI Considerations

- CLI mode: Use carriage return (`\r`) for updating same line
- TUI mode: Integrate with existing Bubble Tea components
- Non-TTY mode: Disable progress indicators automatically
- Windows compatibility: Test with Windows Terminal, CMD, PowerShell

## References

- Original implementation: PR #50
- Suggested by golang-cli-architect for better UX
- Similar to git's progress indicators for large operations