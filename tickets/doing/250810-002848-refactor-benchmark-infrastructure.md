---
priority: 2
description: Create comprehensive benchmark suite for ticketflow refactoring
created_at: "2025-08-10T00:28:48+09:00"
started_at: "2025-08-10T09:57:19+09:00"
closed_at: null
---

# Task 1.1: Benchmark Infrastructure

**Duration**: 0.5 days  
**Complexity**: Low  
**Phase**: 1 - Foundation

Create comprehensive benchmark suite covering all critical paths (list, start, new commands). Establish baseline measurements and set up continuous benchmarking in CI.

## Tasks
Make sure to update task status when you finish it. Also, always create a commit for each task you finished.

- [x] Create benchmark suite using `testing.B` with allocation tracking
- [x] Implement benchmarks for list, start, and new commands  
- [x] Set up comparison benchmarks (sequential vs concurrent)
- [x] Configure benchstat for statistical analysis
- [x] Integrate pprof for CPU and memory profiling
- [x] Establish baseline measurements for all operations
- [x] Set up continuous benchmarking in CI pipeline
- [x] Create benchmark comparison reports
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Update documentation if necessary
- [x] Update README.md
- [x] Update the ticket with insights from resolving this ticket
- [ ] Get developer approval before closing

## Implementation Notes

- Use `b.ReportAllocs()` to track allocations
- Implement `b.StopTimer()`/`b.StartTimer()` for setup exclusion
- Create comparison benchmarks: `BenchmarkListSequential` vs `BenchmarkListConcurrent`
- Use `benchstat` for statistical analysis of results
- Set up `pprof` integration for CPU and memory profiling

## Expected Outcomes

- Comprehensive baseline measurements for all operations
- Automated performance regression detection
- Clear performance targets for optimization work

## Implementation Insights

### What Was Accomplished

1. **Comprehensive Benchmark Suite**: Created advanced benchmarks covering:
   - Various content sizes (100B to 100KB)
   - Large repository scenarios (100, 500, 1000 tickets)
   - Concurrent operations with configurable parallelism (1, 2, 4, 8 workers)
   - Memory pressure testing with simultaneous operations
   - Worktree-specific benchmarks

2. **Benchmark Utilities**: Developed reusable utilities in `internal/testutil/benchmark.go`:
   - `BenchmarkEnvironment` for consistent test setup
   - `BenchmarkTimer` for precise timing control
   - `CreateLargeRepository` for stress testing
   - `MeasureMemoryUsage` for memory profiling

3. **Benchmark Infrastructure**:
   - Shell scripts for running comprehensive and quick benchmarks
   - Baseline comparison tool for regression detection
   - Integration with Make targets for easy execution
   - Support for CPU and memory profiling

### Key Performance Findings

From initial benchmarks on Apple M1 Max:
- **Create Ticket**: ~9.1ms per operation, 57KB allocations, 483 allocs/op
- **List 100 Tickets (Text)**: ~5.9ms, 1.1MB allocations
- **List 100 Tickets (JSON)**: ~12.3ms, 2.7MB allocations (2x slower than text)
- **List 1000 Tickets**: ~58ms (text), ~121ms (JSON)
- **Worktree Creation**: ~137ms per operation (significant overhead)
- **Memory Usage**: Linear scaling with ticket count, JSON format uses 2.4x more memory

### Areas for Future Optimization

1. **JSON Serialization**: Significant overhead compared to text output
2. **Worktree Operations**: High latency due to git operations
3. **Large Repository Listing**: Performance degrades with 1000+ tickets
4. **Memory Allocations**: High allocation count in list operations

### Recommendations

1. Consider caching for frequently accessed ticket lists
2. Implement pagination for large ticket counts
3. Optimize JSON serialization or use more efficient formats
4. Consider batch operations for worktree management
5. Profile and optimize hot paths identified by benchmarks

### Technical Decisions

- Used `b.ReportAllocs()` for all benchmarks to track memory usage
- Implemented `b.StopTimer()`/`b.StartTimer()` for accurate measurements
- Created realistic test scenarios with actual file I/O operations
- Focused on end-to-end benchmarks rather than micro-benchmarks
- Made baseline generation configurable (BENCH_TIME, BENCH_COUNT, BENCH_TIMEOUT)
- Aligned benchtime defaults to 3s for consistency between baseline and comparison

### Implementation Challenges & Solutions

1. **Failing Benchmarks**: Some benchmarks were failing due to git state issues
   - Solution: Added proper git commits and cleanup between benchmark iterations
   - Used force flags for operations that check for uncommitted changes

2. **Benchmark Timeouts**: Large repository tests (1000+ tickets) were timing out
   - Solution: Implemented dynamic timeout calculation based on ticket count
   - Added configurable timeout parameters in Makefile

3. **Inconsistent Benchmark Parameters**: Baseline and comparison used different benchtimes
   - Solution: Made all benchmark parameters configurable via environment variables
   - Standardized on 3s benchtime as default for both

4. **CI Integration**: Ensuring benchmarks work reliably in CI environment
   - Solution: Added proper error handling and made benchmarks more deterministic
   - Fixed all race conditions and timing issues

### Files Created/Modified

**New Files:**
- `internal/cli/commands_advanced_benchmark_test.go` - Advanced benchmark scenarios
- `internal/testutil/benchmark.go` - Reusable benchmark utilities
- `internal/testutil/benchmark_test.go` - Tests for benchmark utilities
- `benchmarks/run-quick.sh` - Quick benchmark execution script
- `benchmarks/run-comprehensive.sh` - Full benchmark suite with profiling
- `benchmarks/compare-with-baseline.sh` - Baseline comparison tool
- `benchmarks/baseline.txt` - Performance baseline data
- `benchmarks/README.md` - Comprehensive documentation

**Modified Files:**
- `internal/cli/commands_benchmark_test.go` - Refactored to use new utilities
- `Makefile` - Added comprehensive benchmark targets
- `.gitignore` - Added benchmark artifacts to ignore list

### PR Status

**PR #49**: https://github.com/yshrsmz/ticketflow/pull/49
- All CI checks passing ✅
- All review comments addressed ✅
- Ready for final review and approval