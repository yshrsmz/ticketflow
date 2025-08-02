---
priority: 2
description: "Implement configurable timeouts for operations"
created_at: "2025-08-02T14:12:19+09:00"
started_at: null
closed_at: null
related:
    - parent:250801-003206-add-context-support
---

# Add Timeout Configuration Support

Implement configurable timeouts for operations to prevent commands from running indefinitely.

## Context

Currently all operations use context.Background() without any timeout. This ticket adds configuration options to set default timeouts for different types of operations, improving reliability and user experience.

## Tasks

- [ ] Add timeout configuration to config.yaml structure
- [ ] Define timeout fields for different operation types (git, file I/O, etc.)
- [ ] Update config package with timeout parsing and validation
- [ ] Modify CLI commands to use timeout from config
- [ ] Add command-line flags to override timeout values
- [ ] Implement graceful timeout handling with proper error messages
- [ ] Add default timeout values (e.g., 30s for git operations)
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update documentation with timeout configuration examples
- [ ] Update README.md with timeout configuration
- [ ] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Configuration Design

```yaml
# .ticketflow.yaml
timeouts:
  git: 30s         # Git operations timeout
  file: 10s        # File I/O operations timeout
  init: 5m         # Init commands timeout
  default: 1m      # Default timeout for other operations
```

## Implementation Notes

1. Use `context.WithTimeout` instead of `context.Background()`
2. Allow per-command timeout overrides via CLI flags
3. Ensure timeout errors are clearly reported to users
4. Consider different timeouts for different git operations (clone vs status)

## Dependencies

- Requires completion of parent ticket: 250801-003206-add-context-support