---
priority: 3
description: "Add progress indicators for long-running operations"
created_at: "2025-08-11T21:54:24+09:00"
started_at: null
closed_at: null
related:
    - parent:250810-002848-refactor-concurrent-directory-ops
---

# Progress Indicators for Large Operations

Add optional progress indicators for operations that take more than 1 second.

## Background

When listing hundreds or thousands of tickets, users have no feedback about operation progress. The golang-cli-architect review suggested adding progress indication for better UX.

## Tasks

- [ ] Design progress indicator interface that works with both CLI and TUI
- [ ] Implement progress bar/spinner for CLI mode
- [ ] Add progress tracking to concurrent operations:
  - Track files processed vs total
  - Show current operation (reading, sorting, etc.)
  - Display elapsed time
- [ ] Make progress indicators configurable:
  - Threshold time before showing (default: 1 second)
  - Enable/disable via config or flag
  - Respect quiet mode (`-q` flag)
- [ ] Ensure progress output goes to stderr (not stdout)
- [ ] Handle terminal capabilities detection
- [ ] Add cancellation hint (e.g., "Press Ctrl+C to cancel")
- [ ] Test with various terminal emulators
- [ ] Document progress indicator behavior
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