#!/bin/bash

# TicketFlow Refactoring - Next Ticket Selector
# This script analyzes refactoring ticket progress and suggests the next ticket to work on

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Ticket phases and dependencies
declare -A TICKET_DEPS=(
    # Phase 1 - Foundation (can parallelize)
    ["refactor-benchmark-infrastructure"]=""
    ["refactor-concurrent-directory-ops"]="refactor-benchmark-infrastructure"
    ["refactor-object-pooling"]="refactor-benchmark-infrastructure"
    ["refactor-parallel-test-execution"]=""
    
    # Phase 2 - Command Architecture
    ["refactor-command-interface"]=""
    ["refactor-command-registry"]="refactor-command-interface"
    ["refactor-worker-pool"]="refactor-command-interface"
    ["refactor-migrate-first-commands"]="refactor-command-registry,refactor-worker-pool"
    
    # Phase 3 - Performance Optimizations
    ["refactor-streaming-architecture"]="refactor-command-registry"
    ["refactor-git-operations-batching"]="refactor-worker-pool"
    ["refactor-configuration-caching"]=""
    ["refactor-yaml-parsing-optimization"]="refactor-object-pooling"
    
    # Phase 4 - Error Handling and Monitoring
    ["refactor-error-aggregation"]="refactor-command-registry"
    ["refactor-circuit-breaker"]="refactor-error-aggregation"
    ["refactor-performance-monitoring"]="refactor-benchmark-infrastructure"
    ["refactor-chaos-testing"]="refactor-worker-pool,refactor-error-aggregation"
    
    # Phase 5 - Migration and Cleanup
    ["refactor-complete-command-migration"]="refactor-command-interface,refactor-command-registry,refactor-worker-pool,refactor-migrate-first-commands"
    ["refactor-remove-legacy-code"]="refactor-complete-command-migration"
    ["refactor-documentation-update"]="refactor-complete-command-migration"
    ["refactor-migration-guide"]="refactor-remove-legacy-code"
)

declare -A TICKET_PHASE=(
    ["refactor-benchmark-infrastructure"]="1"
    ["refactor-concurrent-directory-ops"]="1"
    ["refactor-object-pooling"]="1"
    ["refactor-parallel-test-execution"]="1"
    ["refactor-command-interface"]="2"
    ["refactor-command-registry"]="2"
    ["refactor-worker-pool"]="2"
    ["refactor-migrate-first-commands"]="2"
    ["refactor-streaming-architecture"]="3"
    ["refactor-git-operations-batching"]="3"
    ["refactor-configuration-caching"]="3"
    ["refactor-yaml-parsing-optimization"]="3"
    ["refactor-error-aggregation"]="4"
    ["refactor-circuit-breaker"]="4"
    ["refactor-performance-monitoring"]="4"
    ["refactor-chaos-testing"]="4"
    ["refactor-complete-command-migration"]="5"
    ["refactor-remove-legacy-code"]="5"
    ["refactor-documentation-update"]="5"
    ["refactor-migration-guide"]="5"
)

declare -A TICKET_COMPLEXITY=(
    ["refactor-benchmark-infrastructure"]="Low"
    ["refactor-concurrent-directory-ops"]="Low"
    ["refactor-object-pooling"]="Low"
    ["refactor-parallel-test-execution"]="Low"
    ["refactor-command-interface"]="Medium"
    ["refactor-command-registry"]="Medium"
    ["refactor-worker-pool"]="High"
    ["refactor-migrate-first-commands"]="Medium"
    ["refactor-streaming-architecture"]="High"
    ["refactor-git-operations-batching"]="Medium"
    ["refactor-configuration-caching"]="Low"
    ["refactor-yaml-parsing-optimization"]="Medium"
    ["refactor-error-aggregation"]="Medium"
    ["refactor-circuit-breaker"]="Medium"
    ["refactor-performance-monitoring"]="Low"
    ["refactor-chaos-testing"]="High"
    ["refactor-complete-command-migration"]="Medium"
    ["refactor-remove-legacy-code"]="Low"
    ["refactor-documentation-update"]="Low"
    ["refactor-migration-guide"]="Low"
)

declare -A TICKET_DURATION=(
    ["refactor-benchmark-infrastructure"]="0.5 days"
    ["refactor-concurrent-directory-ops"]="1 day"
    ["refactor-object-pooling"]="0.5 days"
    ["refactor-parallel-test-execution"]="0.5 days"
    ["refactor-command-interface"]="1 day"
    ["refactor-command-registry"]="2 days"
    ["refactor-worker-pool"]="2 days"
    ["refactor-migrate-first-commands"]="2 days"
    ["refactor-streaming-architecture"]="2 days"
    ["refactor-git-operations-batching"]="1 day"
    ["refactor-configuration-caching"]="1 day"
    ["refactor-yaml-parsing-optimization"]="1 day"
    ["refactor-error-aggregation"]="1 day"
    ["refactor-circuit-breaker"]="1 day"
    ["refactor-performance-monitoring"]="1 day"
    ["refactor-chaos-testing"]="1.5 days"
    ["refactor-complete-command-migration"]="3 days"
    ["refactor-remove-legacy-code"]="1 day"
    ["refactor-documentation-update"]="1 day"
    ["refactor-migration-guide"]="0.5 days"
)

# Function to get ticket status
get_ticket_status() {
    local ticket_name=$1
    
    # Check in done directory
    if ls tickets/done/*-${ticket_name}.md 2>/dev/null | grep -q .; then
        echo "done"
        return
    fi
    
    # Check in doing directory
    if ls tickets/doing/*-${ticket_name}.md 2>/dev/null | grep -q .; then
        echo "doing"
        return
    fi
    
    # Check in todo directory
    if ls tickets/todo/*-${ticket_name}.md 2>/dev/null | grep -q .; then
        echo "todo"
        return
    fi
    
    echo "not_found"
}

# Function to check if all dependencies are met
check_dependencies() {
    local ticket=$1
    local deps="${TICKET_DEPS[$ticket]}"
    
    if [ -z "$deps" ]; then
        return 0  # No dependencies
    fi
    
    IFS=',' read -ra DEP_ARRAY <<< "$deps"
    for dep in "${DEP_ARRAY[@]}"; do
        local dep_status=$(get_ticket_status "$dep")
        if [ "$dep_status" != "done" ]; then
            return 1  # Dependency not met
        fi
    done
    
    return 0  # All dependencies met
}

# Function to get ticket file path
get_ticket_path() {
    local ticket_name=$1
    local path=$(ls tickets/todo/*-${ticket_name}.md 2>/dev/null | head -n1)
    if [ -z "$path" ]; then
        path=$(ls tickets/doing/*-${ticket_name}.md 2>/dev/null | head -n1)
    fi
    if [ -z "$path" ]; then
        path=$(ls tickets/done/*-${ticket_name}.md 2>/dev/null | head -n1)
    fi
    echo "$path"
}

# Function to display ticket context
display_ticket_context() {
    local ticket=$1
    local ticket_path=$(get_ticket_path "$ticket")
    
    echo -e "\n${BOLD}${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${BOLD}${GREEN}🎯 Selected Ticket: ${ticket}${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
    
    # Display ticket metadata
    echo -e "${BOLD}📊 Metadata:${NC}"
    echo -e "  ${BLUE}Phase:${NC} ${TICKET_PHASE[$ticket]}"
    echo -e "  ${BLUE}Complexity:${NC} ${TICKET_COMPLEXITY[$ticket]}"
    echo -e "  ${BLUE}Duration:${NC} ${TICKET_DURATION[$ticket]}"
    echo -e "  ${BLUE}Dependencies:${NC} ${TICKET_DEPS[$ticket]:-None}"
    
    # Display ticket content
    if [ -f "$ticket_path" ]; then
        echo -e "\n${BOLD}📄 Ticket Content:${NC}"
        echo -e "${CYAN}────────────────────────────────────────────────────────────${NC}"
        cat "$ticket_path" | sed 's/^/  /'
        echo -e "${CYAN}────────────────────────────────────────────────────────────${NC}"
    fi
    
    # Display relevant documentation
    echo -e "\n${BOLD}📚 Related Documentation:${NC}"
    echo -e "  • ${BLUE}Full refactoring discussion:${NC} docs/20250810-refactor-discussion.md"
    echo -e "  • ${BLUE}Executive summary:${NC} docs/20250810-refactor-summary.md"
    echo -e "  • ${BLUE}Ticket overview:${NC} docs/20250810-refactor-tickets.md"
    
    # Display specific implementation notes from summary
    echo -e "\n${BOLD}💡 Implementation Notes:${NC}"
    case "$ticket" in
        "refactor-benchmark-infrastructure")
            echo "  • Use testing.B with b.ReportAllocs() to track allocations"
            echo "  • Implement b.StopTimer()/b.StartTimer() for setup exclusion"
            echo "  • Create comparison benchmarks: BenchmarkListSequential vs BenchmarkListConcurrent"
            echo "  • Use benchstat for statistical analysis"
            echo "  • Set up pprof integration for CPU and memory profiling"
            echo "  • Key files: Create benchmark_test.go files in relevant packages"
            echo "  • Reference: golang.org/x/perf/cmd/benchstat"
            ;;
        "refactor-concurrent-directory-ops")
            echo "  • Use errgroup.Group for structured concurrency with error propagation"
            echo "  • Implement semaphore.NewWeighted() to limit concurrent file operations"
            echo "  • Pre-allocate result slices with estimated capacity"
            echo "  • Use runtime.NumCPU() to determine optimal worker count"
            echo "  • Implement context cancellation checks in tight loops"
            echo "  • Key files: internal/ticket/manager.go (List method)"
            echo "  • Reference: golang.org/x/sync/errgroup, golang.org/x/sync/semaphore"
            ;;
        "refactor-object-pooling")
            echo "  • Initialize sync.Pool with factory function that pre-allocates slice capacity"
            echo "  • Clear all references before returning objects to pool to avoid memory leaks"
            echo "  • Use separate pools for different object sizes"
            echo "  • Benchmark allocation rate with runtime.MemStats before/after pooling"
            echo "  • Key files: internal/ticket/ticket.go, internal/ticket/manager.go"
            echo "  • Common pattern: sync.Pool{New: func() interface{} { return &Ticket{} }}"
            ;;
        "refactor-worker-pool")
            echo "  • Pre-allocate command channels with 2x worker count for buffering"
            echo "  • Implement panic recovery in each worker with stack trace capture"
            echo "  • Use runtime.LockOSThread() for CPU affinity in performance-critical workers"
            echo "  • Add circuit breaker pattern with failure threshold and reset timeout"
            echo "  • Monitor goroutine count with runtime.NumGoroutine() for leak detection"
            echo "  • Key files: Create new package internal/worker/"
            echo "  • Reference implementation: See docs/20250810-refactor-discussion.md for code examples"
            ;;
        "refactor-streaming-architecture")
            echo "  • Use bounded channels (capacity 100-1000) for backpressure"
            echo "  • Implement batch reading with timeout for efficiency"
            echo "  • Monitor channel pressure with atomic counters"
            echo "  • Add metrics for processed vs dropped items"
            echo "  • Use io.Pipe() for zero-copy streaming where possible"
            echo "  • Key files: Create internal/stream/ package"
            echo "  • Pattern: Producer-consumer with buffered channels"
            ;;
        "refactor-command-interface")
            echo "  • Define Command interface with Execute, Metadata, and ValidationRules methods"
            echo "  • Support both sync and async execution modes"
            echo "  • Include performance hints in metadata"
            echo "  • Key files: Create internal/command/interface.go"
            echo "  • Pattern: Strategy pattern with command metadata"
            ;;
        "refactor-command-registry")
            echo "  • Implement self-registering commands using init() functions"
            echo "  • Build registry map[string]Command for O(1) lookup"
            echo "  • Support command aliases and help generation"
            echo "  • Key files: internal/command/registry.go, cmd/ticketflow/main.go"
            echo "  • Reference: How git handles command registration"
            ;;
        "refactor-parallel-test-execution")
            echo "  • Add t.Parallel() to all unit tests that don't share state"
            echo "  • Document tests that cannot be parallelized (integration tests)"
            echo "  • Run tests with -race flag to detect issues"
            echo "  • Key files: All *_test.go files except integration tests"
            echo "  • Note: Integration tests use os.Chdir and cannot be parallelized"
            ;;
        "refactor-configuration-caching")
            echo "  • Implement TTL-based cache with sync.RWMutex for thread safety"
            echo "  • Cache parsed YAML configuration for 5 minutes"
            echo "  • Invalidate cache on file modification time change"
            echo "  • Key files: internal/config/config.go"
            echo "  • Pattern: Lazy loading with expiration"
            ;;
        "refactor-yaml-parsing-optimization")
            echo "  • Reuse buffers for YAML parsing with sync.Pool"
            echo "  • Stream large YAML files instead of loading entirely into memory"
            echo "  • Pre-compile regex patterns used in parsing"
            echo "  • Key files: internal/ticket/ticket.go (ParseYAML methods)"
            echo "  • Consider: gopkg.in/yaml.v3 streaming decoder"
            ;;
        "refactor-git-operations-batching")
            echo "  • Batch multiple git commands into single exec calls where possible"
            echo "  • Use git plumbing commands for better performance"
            echo "  • Implement concurrent git operations for independent queries"
            echo "  • Key files: internal/git/*.go"
            echo "  • Example: git for-each-ref instead of multiple git branch calls"
            ;;
        "refactor-error-aggregation")
            echo "  • Categorize errors: Critical, Retryable, Warning"
            echo "  • Implement error collector for concurrent operations"
            echo "  • Provide structured error messages with context"
            echo "  • Key files: Create internal/errors/ package"
            echo "  • Pattern: Multi-error with categorization"
            ;;
        "refactor-circuit-breaker")
            echo "  • Implement three states: Closed, Open, Half-Open"
            echo "  • Configure failure threshold and reset timeout"
            echo "  • Add exponential backoff for retries"
            echo "  • Key files: internal/resilience/circuit_breaker.go"
            echo "  • Reference: sony/gobreaker or similar patterns"
            ;;
        "refactor-performance-monitoring")
            echo "  • Add metrics for command execution time"
            echo "  • Track memory allocations and GC pressure"
            echo "  • Export metrics in OpenTelemetry format"
            echo "  • Key files: internal/metrics/ package"
            echo "  • Consider: prometheus client or OpenTelemetry SDK"
            ;;
        "refactor-chaos-testing")
            echo "  • Inject random delays and failures in tests"
            echo "  • Test concurrent operations under stress"
            echo "  • Verify no goroutine leaks or deadlocks"
            echo "  • Key files: test/chaos/ directory"
            echo "  • Tools: go test -race, goleak for goroutine leak detection"
            ;;
        "refactor-migrate-first-commands")
            echo "  • Start with list, new, and start commands"
            echo "  • Implement using new Command interface"
            echo "  • Register in command registry"
            echo "  • Key files: internal/cli/{list,new,start}.go"
            echo "  • Maintain backward compatibility during migration"
            ;;
        "refactor-complete-command-migration")
            echo "  • Migrate all remaining commands to new architecture"
            echo "  • Ensure consistent error handling across commands"
            echo "  • Update help text and documentation"
            echo "  • Key files: All files in internal/cli/"
            echo "  • Test each command thoroughly after migration"
            ;;
        "refactor-remove-legacy-code")
            echo "  • Remove old switch statement from main.go"
            echo "  • Delete unused handler functions"
            echo "  • Clean up deprecated interfaces"
            echo "  • Key files: cmd/ticketflow/main.go, internal/cli/"
            echo "  • Ensure no dead code remains"
            ;;
        "refactor-documentation-update")
            echo "  • Update architecture documentation"
            echo "  • Document new patterns and best practices"
            echo "  • Add performance tuning guide"
            echo "  • Key files: docs/, README.md, CLAUDE.md"
            echo "  • Include code examples and benchmarks"
            ;;
        "refactor-migration-guide")
            echo "  • Document any breaking changes"
            echo "  • Provide upgrade instructions"
            echo "  • List new features and improvements"
            echo "  • Key files: docs/MIGRATION.md"
            echo "  • Include before/after examples"
            ;;
    esac
    
    # Display quick start commands
    echo -e "\n${BOLD}🚀 Quick Start Commands:${NC}"
    local ticket_id=$(basename "$ticket_path" .md)
    echo -e "  ${GREEN}# Start working on this ticket:${NC}"
    echo -e "  ticketflow start $ticket_id"
    echo -e "  cd ../ticketflow.worktrees/$ticket_id"
    echo
    echo -e "  ${GREEN}# View the ticket:${NC}"
    echo -e "  ticketflow show $ticket_id"
    echo
    echo -e "  ${GREEN}# After completing work:${NC}"
    echo -e "  ticketflow close  # Run from within the worktree"
    echo -e "  git push"
}

# Main function to analyze and select next ticket
analyze_refactor_progress() {
    echo -e "${BOLD}${CYAN}TicketFlow Refactoring Progress Analyzer${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}\n"
    
    # Analyze ticket status
    local todo_tickets=()
    local doing_tickets=()
    local done_tickets=()
    local available_tickets=()
    
    for ticket in "${!TICKET_DEPS[@]}"; do
        local status=$(get_ticket_status "$ticket")
        case "$status" in
            "todo")
                todo_tickets+=("$ticket")
                if check_dependencies "$ticket"; then
                    available_tickets+=("$ticket")
                fi
                ;;
            "doing")
                doing_tickets+=("$ticket")
                ;;
            "done")
                done_tickets+=("$ticket")
                ;;
        esac
    done
    
    # Display progress summary
    echo -e "${BOLD}📈 Progress Summary:${NC}"
    echo -e "  ${GREEN}✓ Completed:${NC} ${#done_tickets[@]} tickets"
    echo -e "  ${YELLOW}⚡ In Progress:${NC} ${#doing_tickets[@]} tickets"
    echo -e "  ${BLUE}○ Todo:${NC} ${#todo_tickets[@]} tickets"
    echo -e "  ${CYAN}◆ Available (deps met):${NC} ${#available_tickets[@]} tickets"
    
    # Display tickets in progress
    if [ ${#doing_tickets[@]} -gt 0 ]; then
        echo -e "\n${BOLD}${YELLOW}⚡ Currently In Progress:${NC}"
        for ticket in "${doing_tickets[@]}"; do
            echo -e "  • $ticket (Phase ${TICKET_PHASE[$ticket]}, ${TICKET_COMPLEXITY[$ticket]} complexity)"
        done
    fi
    
    # Select next ticket
    if [ ${#available_tickets[@]} -eq 0 ]; then
        if [ ${#todo_tickets[@]} -eq 0 ]; then
            echo -e "\n${GREEN}🎉 All refactoring tickets are complete!${NC}"
        else
            echo -e "\n${YELLOW}⚠️  No tickets available - waiting for dependencies:${NC}"
            for ticket in "${todo_tickets[@]}"; do
                echo -e "  • $ticket - waiting for: ${TICKET_DEPS[$ticket]}"
            done
        fi
        return
    fi
    
    # Sort available tickets by phase and select the best one
    local selected_ticket=""
    local min_phase=999
    
    for ticket in "${available_tickets[@]}"; do
        local phase="${TICKET_PHASE[$ticket]}"
        if [ "$phase" -lt "$min_phase" ]; then
            min_phase=$phase
            selected_ticket=$ticket
        fi
    done
    
    # If multiple tickets in same phase, prefer low complexity ones
    if [ -n "$selected_ticket" ]; then
        for ticket in "${available_tickets[@]}"; do
            if [ "${TICKET_PHASE[$ticket]}" == "$min_phase" ] && [ "${TICKET_COMPLEXITY[$ticket]}" == "Low" ]; then
                selected_ticket=$ticket
                break
            fi
        done
    fi
    
    # Display recommendation
    display_ticket_context "$selected_ticket"
    
    # Show other available options
    if [ ${#available_tickets[@]} -gt 1 ]; then
        echo -e "\n${BOLD}📋 Other Available Tickets:${NC}"
        for ticket in "${available_tickets[@]}"; do
            if [ "$ticket" != "$selected_ticket" ]; then
                echo -e "  • $ticket (Phase ${TICKET_PHASE[$ticket]}, ${TICKET_COMPLEXITY[$ticket]} complexity, ${TICKET_DURATION[$ticket]})"
            fi
        done
    fi
}

# Run the analyzer
analyze_refactor_progress