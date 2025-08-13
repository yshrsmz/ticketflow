---
priority: 4
description: "Add performance metrics collection and reporting"
created_at: "2025-08-11T21:53:50+09:00"
started_at: null
closed_at: null
related:
    - parent:250810-002848-refactor-concurrent-directory-ops
---

# Performance Metrics Instrumentation

Add instrumentation to collect and report performance metrics for concurrent operations.

## Background

The concurrent directory operations implementation provides debug logging, but lacks detailed metrics collection for production monitoring and performance analysis.

## Tasks

- [ ] Design metrics collection interface
- [ ] Implement metrics collector for:
  - Operation duration (list, create, update, etc.)
  - Concurrency level used
  - Queue wait times (semaphore acquisition)
  - File I/O statistics
  - Cache hit/miss rates (if caching is implemented)
- [ ] Add OpenTelemetry integration (optional)
- [ ] Create metrics reporting commands:
  - `ticketflow metrics show` - Display current session metrics
  - `ticketflow metrics reset` - Reset metrics counters
- [ ] Add metrics export formats:
  - JSON for programmatic access
  - Human-readable summary
  - Prometheus format (optional)
- [ ] Document metrics collection and usage
- [ ] Add benchmarking mode that collects detailed metrics
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Update README.md
- [ ] Get developer approval before closing

## Implementation Notes

- Metrics collection should be zero-overhead when disabled
- Use atomic operations for thread-safe counters
- Consider using `expvar` package for runtime metrics
- Metrics should be opt-in via configuration

## Example Metrics Output

```
TicketFlow Performance Metrics
==============================
List Operations:
  Total: 42
  Sequential: 12 (avg: 5.2ms)
  Concurrent: 30 (avg: 2.1ms)
  
Concurrency:
  Avg Workers: 6.2
  Max Workers: 8
  Semaphore Wait: 0.3ms avg
  
File I/O:
  Files Read: 1,234
  Bytes Read: 456KB
  Avg Read Time: 1.2ms
```

## References

- Original implementation: PR #50
- Suggested by golang-pro review for production observability
- Could integrate with existing monitoring systems