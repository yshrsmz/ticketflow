---
priority: 2
description: Fix cleanup command ignoring --format parameter
created_at: "2025-08-20T13:57:14+09:00"
started_at: "2025-08-20T14:00:33+09:00"
closed_at: null
---

# Fix cleanup command ignoring --format parameter

## Problem
The `ticketflow cleanup` command doesn't respect the `--format` parameter. When running `ticketflow cleanup --format json`, it still outputs human-readable format instead of JSON.

## Expected Behavior
When `--format json` is specified, the cleanup command should output JSON format for consistency with other commands and better AI/tool integration.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [ ] Investigate cleanup command implementation in `internal/cli/commands/cleanup.go`
- [ ] Add format parameter handling to cleanup command
- [ ] Ensure JSON output follows same structure as other commands
- [ ] Add tests for JSON output format
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation if necessary
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Notes

This affects AI integration as JSON format is preferred for structured parsing of command outputs.