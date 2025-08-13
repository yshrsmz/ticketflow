# TicketFlow Benchmarks

This directory contains a simple performance benchmark for TicketFlow's most critical operation.

## Running Benchmarks

Run the benchmark manually when curious about performance:

```bash
# Run the List operation benchmark
go test -bench=BenchmarkManagerList -benchmem ./internal/ticket

# Run with longer duration for more accurate results
go test -bench=BenchmarkManagerList -benchmem -benchtime=10s ./internal/ticket
```

## Current Performance

The List operation (most critical path) performs excellently:
- 10 tickets: ~0.4ms
- 50 tickets: ~1.4ms  
- 100 tickets: ~2.7ms

These times are already very fast for a CLI tool. The concurrent implementation automatically kicks in for larger ticket counts.

## Note

Benchmarks are kept minimal and run manually only. They are not part of CI to avoid unnecessary overhead. TicketFlow's performance is already excellent for its use case as a local developer tool.