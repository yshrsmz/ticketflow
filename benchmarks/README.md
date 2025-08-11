# TicketFlow Benchmarks

This directory contains performance benchmarks for the TicketFlow project. These benchmarks help track performance characteristics and identify optimization opportunities.

## Running Benchmarks

### Quick Start

Run all benchmarks:
```bash
make bench
```

Run specific package benchmarks:
```bash
make bench-cli      # CLI command benchmarks
make bench-ticket   # Ticket manager benchmarks
make bench-git      # Git operations benchmarks
```

### Advanced Usage

Run benchmarks with specific patterns:
```bash
go test -bench=BenchmarkList ./internal/cli -benchmem
go test -bench=BenchmarkCreate ./internal/ticket -benchmem
```

Run benchmarks with CPU profiling:
```bash
go test -bench=. ./internal/cli -cpuprofile=cpu.prof -benchmem
go tool pprof cpu.prof
```

Run benchmarks with memory profiling:
```bash
go test -bench=. ./internal/ticket -memprofile=mem.prof -benchmem
go tool pprof mem.prof
```

### Benchmark Options

- `-benchtime=10s`: Run each benchmark for at least 10 seconds
- `-benchmem`: Include memory allocation statistics
- `-count=5`: Run each benchmark 5 times for more reliable results
- `-cpu=1,2,4`: Run benchmarks with different GOMAXPROCS values
- `-run=^$`: Skip regular tests, only run benchmarks

### Running Comprehensive Benchmarks

For thorough performance testing:
```bash
./benchmarks/run-comprehensive.sh
```

This script:
- Runs all benchmarks multiple times
- Captures CPU and memory profiles
- Generates comparison reports
- Saves results with timestamps

## Benchmark Categories

### 1. CLI Command Benchmarks (`internal/cli`)

- **BenchmarkCreateTicket**: Ticket creation performance
- **BenchmarkStartTicket**: Starting tickets with/without worktrees
- **BenchmarkCloseTicket**: Closing tickets
- **BenchmarkListTickets**: Listing tickets with various filters and formats
- **BenchmarkNewTicketWithVariousSizes**: Creation with different content sizes
- **BenchmarkListTicketsLargeRepository**: Performance with large ticket counts
- **BenchmarkStartTicketConcurrent**: Concurrent ticket operations
- **BenchmarkCloseTicketWithReason**: Close operations with reasons
- **BenchmarkWorktreeOperations**: Worktree-specific operations
- **BenchmarkSearchAndFilter**: Search and filtering performance
- **BenchmarkMemoryPressure**: Performance under memory pressure

### 2. Ticket Manager Benchmarks (`internal/ticket`)

- **BenchmarkManagerCreate**: Core ticket creation
- **BenchmarkManagerGet**: Single ticket retrieval
- **BenchmarkManagerList**: Listing with different filters
- **BenchmarkManagerUpdate**: Ticket updates
- **BenchmarkManagerFindTicket**: Ticket search operations
- **BenchmarkManagerReadWriteContent**: File I/O operations
- **BenchmarkManagerCreateConcurrent**: Concurrent creation
- **BenchmarkManagerListConcurrent**: Concurrent listing
- **BenchmarkManagerCurrentTicket**: Current ticket operations

### 3. Git Operations Benchmarks (`internal/git`)

- **BenchmarkGitExec**: Basic git command execution
- **BenchmarkCreateBranch**: Branch creation
- **BenchmarkBranchExists**: Branch existence checks
- **BenchmarkCurrentBranch**: Getting current branch
- **BenchmarkListWorktrees**: Worktree listing
- **BenchmarkAddWorktree**: Worktree creation
- **BenchmarkCommit**: Commit operations

## CI Integration

Pull requests automatically trigger a dedicated "Benchmark Check" job that:
- Runs as a separate GitHub Actions job for better visibility
- Uses `BENCH_TIME=100ms` for fast feedback (~30 seconds)
- Runs only on core packages (cli, ticket, git) for speed
- Shows warnings for potential regressions but doesn't block merging
- Uploads benchmark results as artifacts for detailed review
- Runs in parallel with other CI checks, not blocking critical tests

For comprehensive benchmarking, run locally before merging significant changes:
```bash
make bench-compare  # Full comparison with baseline
```

## Performance Baselines

Current performance baselines (as of latest run):

### Key Operations

| Operation | Time/op | Allocations/op | Bytes/op |
|-----------|---------|----------------|----------|
| Create Ticket | ~X ms | ~Y allocs | ~Z B |
| List 100 Tickets (Text) | ~X ms | ~Y allocs | ~Z B |
| List 100 Tickets (JSON) | ~X ms | ~Y allocs | ~Z B |
| Start Ticket (no worktree) | ~X ms | ~Y allocs | ~Z B |
| Start Ticket (with worktree) | ~X ms | ~Y allocs | ~Z B |

*Note: Update these values after running `make bench-baseline`*

## Continuous Benchmarking

### Comparing Performance

Compare current performance with baseline:
```bash
./benchmarks/compare-with-baseline.sh
```

### Regression Detection

The CI pipeline automatically runs benchmarks and compares with the baseline. Performance regressions exceeding thresholds will fail the build.

Thresholds:
- Time regression: >10% slower
- Memory regression: >20% more allocations
- Critical operations (Create, List, Start): >5% slower

## Optimization Guidelines

When optimizing based on benchmark results:

1. **Profile First**: Use CPU and memory profiling to identify hotspots
2. **Measure Impact**: Run benchmarks before and after changes
3. **Consider Trade-offs**: Balance speed, memory usage, and code complexity
4. **Document Changes**: Note optimizations in commit messages
5. **Update Baselines**: After significant improvements, update baseline.txt

## Writing New Benchmarks

When adding new benchmarks:

1. Use the benchmark utilities in `internal/testutil/benchmark.go`
2. Include `b.ReportAllocs()` for memory statistics
3. Use `b.StopTimer()`/`b.StartTimer()` for setup/teardown
4. Create realistic test scenarios
5. Add documentation here for new benchmark categories

Example:
```go
func BenchmarkNewFeature(b *testing.B) {
    env := testutil.SetupBenchmarkEnvironment(b)
    
    // Setup
    b.StopTimer()
    // ... prepare test data
    b.StartTimer()
    
    b.ResetTimer()
    b.ReportAllocs()
    
    for i := 0; i < b.N; i++ {
        // Benchmark operation
    }
}
```

## Troubleshooting

### Common Issues

1. **Benchmarks taking too long**: Reduce `-benchtime` or use pattern matching
2. **Inconsistent results**: Increase `-count` or close other applications
3. **Out of memory**: Reduce concurrent operations or test data size
4. **Git errors**: Ensure clean git state before running benchmarks

### Debug Mode

Enable verbose output:
```bash
go test -bench=. -v ./internal/cli
```

## Related Documentation

- [Testing Patterns](../docs/testing-patterns.md)
- [Performance Monitoring](../docs/performance-monitoring.md)
- [Context Usage](../docs/context-usage.md)