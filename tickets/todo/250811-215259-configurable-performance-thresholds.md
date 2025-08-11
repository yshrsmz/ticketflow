---
priority: 3
description: "Add configuration options for performance tuning"
created_at: "2025-08-11T21:52:59+09:00"
started_at: null
closed_at: null
related:
    - parent:250810-002848-refactor-concurrent-directory-ops
---

# Configurable Performance Thresholds

Add configuration options to allow users to tune performance parameters for concurrent operations.

## Background

The concurrent directory operations implementation (Task 1.2) uses hardcoded values:
- Concurrency threshold: 10 files
- Maximum workers: 8

Power users may want to tune these based on their system characteristics and workload.

## Tasks

- [ ] Add performance configuration section to `.ticketflow.yaml` schema
- [ ] Add config fields for:
  - `concurrent_threshold` (default: 10)
  - `max_workers` (default: 8)
  - `enable_concurrent` (default: true) - to optionally disable
- [ ] Update `internal/config` to parse these settings
- [ ] Modify `internal/ticket/manager.go` to use config values instead of constants
- [ ] Add validation for config values (min/max bounds)
- [ ] Update documentation with configuration examples
- [ ] Add tests for configuration parsing and validation
- [ ] Consider environment variable overrides (e.g., `TICKETFLOW_MAX_WORKERS`)
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update README.md
- [ ] Get developer approval before closing

## Implementation Notes

- Keep zero-config default behavior - all settings should be optional
- Use sensible defaults that work for most users
- Validate bounds (e.g., max_workers should be 1-100)
- Log configuration values at debug level for troubleshooting

## Example Configuration

```yaml
# .ticketflow.yaml
performance:
  concurrent_threshold: 10  # Min files before using concurrent loading
  max_workers: 8            # Max concurrent workers
  enable_concurrent: true   # Enable/disable concurrent operations
```

## References

- Original implementation: PR #50
- Code review suggestion from golang-cli-architect
- Related to refactoring phase 1 improvements