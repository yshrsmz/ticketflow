#!/bin/bash

# run-comprehensive.sh - Run comprehensive benchmarks for TicketFlow
set -euo pipefail

# Check for required commands
command -v go >/dev/null 2>&1 || { echo "Error: 'go' command not found. Please install Go."; exit 1; }
command -v grep >/dev/null 2>&1 || { echo "Error: 'grep' command not found."; exit 1; }
command -v awk >/dev/null 2>&1 || { echo "Error: 'awk' command not found."; exit 1; }

# Colors for output (with color support detection)
if [ -t 1 ] && [ "${TERM:-}" != "dumb" ] && command -v tput >/dev/null 2>&1 && [ "$(tput colors 2>/dev/null || echo 0)" -ge 8 ]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m' # No Color
else
    RED=''
    GREEN=''
    YELLOW=''
    NC=''
fi

# Configuration
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
RESULTS_DIR="benchmarks/results/${TIMESTAMP}"
BASELINE_FILE="benchmarks/baseline.txt"
BENCH_TIME="${BENCH_TIME:-10s}"
BENCH_COUNT="${BENCH_COUNT:-3}"
REGRESSION_THRESHOLD="${REGRESSION_THRESHOLD:-10}"  # Default 10% regression threshold

# Create results directory
mkdir -p "${RESULTS_DIR}"

echo -e "${GREEN}Starting comprehensive benchmark suite...${NC}"
echo "Timestamp: ${TIMESTAMP}"
echo "Results directory: ${RESULTS_DIR}"
echo "Benchmark time: ${BENCH_TIME}"
echo "Benchmark count: ${BENCH_COUNT}"
echo

# Function to run benchmarks for a package
run_package_benchmarks() {
    local package=$1
    local name=$2
    local output_file="${RESULTS_DIR}/${name}.txt"
    local cpu_profile="${RESULTS_DIR}/${name}_cpu.prof"
    local mem_profile="${RESULTS_DIR}/${name}_mem.prof"
    
    echo -e "${YELLOW}Running ${name} benchmarks...${NC}"
    
    # Run benchmarks with profiling
    go test -bench=. \
        -benchmem \
        -benchtime="${BENCH_TIME}" \
        -count="${BENCH_COUNT}" \
        -cpuprofile="${cpu_profile}" \
        -memprofile="${mem_profile}" \
        -run=^$ \
        "${package}" > "${output_file}" 2>&1
    
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ ${name} benchmarks completed${NC}"
        
        # Extract summary statistics
        echo "  Results summary:"
        grep "Benchmark" "${output_file}" | head -5 | while read line; do
            echo "    ${line}"
        done
    else
        echo -e "${RED}✗ ${name} benchmarks failed${NC}"
        cat "${output_file}"
    fi
    echo
}

# Run benchmarks for each package
echo -e "${GREEN}=== CLI Command Benchmarks ===${NC}"
run_package_benchmarks "./internal/cli" "cli_commands"

echo -e "${GREEN}=== Ticket Manager Benchmarks ===${NC}"
run_package_benchmarks "./internal/ticket" "ticket_manager"

echo -e "${GREEN}=== Git Operations Benchmarks ===${NC}"
run_package_benchmarks "./internal/git" "git_operations"

echo -e "${GREEN}=== UI/TUI Benchmarks ===${NC}"
run_package_benchmarks "./internal/ui" "ui_tui"

# Combine all results
echo -e "${YELLOW}Combining results...${NC}"
cat "${RESULTS_DIR}"/*.txt > "${RESULTS_DIR}/all_benchmarks.txt"

# Generate summary report
echo -e "${YELLOW}Generating summary report...${NC}"
cat > "${RESULTS_DIR}/summary.md" << EOF
# Benchmark Results - ${TIMESTAMP}

## Configuration
- Benchmark Time: ${BENCH_TIME}
- Benchmark Count: ${BENCH_COUNT}
- Go Version: $(go version)
- Platform: $(uname -s) $(uname -m)

## Top Performance Metrics

### CLI Commands
$(grep -E "^Benchmark.*ns/op" "${RESULTS_DIR}/cli_commands.txt" | head -10 || echo "No results")

### Ticket Manager
$(grep -E "^Benchmark.*ns/op" "${RESULTS_DIR}/ticket_manager.txt" | head -10 || echo "No results")

### Git Operations
$(grep -E "^Benchmark.*ns/op" "${RESULTS_DIR}/git_operations.txt" | head -10 || echo "No results")

## Memory Allocation Summary

### Highest Allocations
$(grep -E "allocs/op" "${RESULTS_DIR}/all_benchmarks.txt" | sort -t' ' -k5 -rn | head -10 || echo "No results")

## Profiles Generated
- CPU Profiles: ${RESULTS_DIR}/*_cpu.prof
- Memory Profiles: ${RESULTS_DIR}/*_mem.prof

To analyze profiles:
\`\`\`bash
go tool pprof ${RESULTS_DIR}/cli_commands_cpu.prof
go tool pprof ${RESULTS_DIR}/ticket_manager_mem.prof
\`\`\`
EOF

echo -e "${GREEN}Summary report generated: ${RESULTS_DIR}/summary.md${NC}"

# Compare with baseline if it exists
if [ -f "${BASELINE_FILE}" ]; then
    echo
    echo -e "${YELLOW}Comparing with baseline...${NC}"
    echo "Using regression threshold: ${REGRESSION_THRESHOLD}%"
    
    # Simple comparison - check if any benchmarks got significantly slower
    # This is a basic implementation - could be enhanced with proper statistical analysis
    
    echo "Checking for performance regressions..."
    
    # Extract benchmark names and times from current run
    grep -E "^Benchmark" "${RESULTS_DIR}/all_benchmarks.txt" | while read line; do
        bench_name=$(echo "$line" | awk '{print $1}')
        current_time=$(echo "$line" | awk '{print $3}')
        
        # Find same benchmark in baseline using awk for safer matching
        baseline_line=$(awk -v bench="${bench_name}" '$1 == bench' "${BASELINE_FILE}" 2>/dev/null || true)
        
        if [ -n "${baseline_line}" ]; then
            baseline_time=$(echo "${baseline_line}" | awk '{print $3}')
            
            # Calculate percentage change using awk for better portability
            # Extract numeric values (remove ns/op suffix and handle scientific notation)
            current_val=$(echo "${current_time}" | sed 's/ns\/op//')
            baseline_val=$(echo "${baseline_time}" | sed 's/ns\/op//')
            
            # Use awk for floating point arithmetic (more portable than bc)
            # Check if baseline_val exists and is not zero using awk for numeric comparison
            if [ -n "${baseline_val}" ] && [ "$(awk -v val="${baseline_val}" 'BEGIN { print (val != 0) ? 1 : 0 }')" -eq 1 ]; then
                result=$(awk -v curr="${current_val}" -v base="${baseline_val}" -v thresh="${REGRESSION_THRESHOLD}" '
                BEGIN {
                    if (base != 0) {
                        change = ((curr - base) / base) * 100
                        printf "%.2f", change
                    } else {
                        print "0"
                    }
                }')
                
                # Use awk for comparison as well
                regression_check=$(awk -v change="${result}" -v thresh="${REGRESSION_THRESHOLD}" 'BEGIN { print (change > thresh) ? 1 : 0 }')
                improvement_check=$(awk -v change="${result}" -v thresh="${REGRESSION_THRESHOLD}" 'BEGIN { print (change < -thresh) ? 1 : 0 }')
                
                if [ "${regression_check}" -eq 1 ]; then
                    echo -e "  ${bench_name}: current=${current_time} baseline=${baseline_time} ${RED}(+${result}% REGRESSION)${NC}"
                elif [ "${improvement_check}" -eq 1 ]; then
                    echo -e "  ${bench_name}: current=${current_time} baseline=${baseline_time} ${GREEN}(${result}% improvement)${NC}"
                else
                    echo "  ${bench_name}: current=${current_time} baseline=${baseline_time} (${result}%)"
                fi
            else
                echo "  ${bench_name}: current=${current_time} baseline=${baseline_time}"
            fi
        fi
    done
else
    echo
    echo -e "${YELLOW}No baseline found. Creating baseline...${NC}"
    cp "${RESULTS_DIR}/all_benchmarks.txt" "${BASELINE_FILE}"
    echo -e "${GREEN}Baseline created: ${BASELINE_FILE}${NC}"
fi

# Final summary
echo
echo -e "${GREEN}=== Benchmark Suite Complete ===${NC}"
echo "Results saved to: ${RESULTS_DIR}"
echo "Summary report: ${RESULTS_DIR}/summary.md"
echo
echo "To view CPU profile:"
echo "  go tool pprof -http=:8080 ${RESULTS_DIR}/cli_commands_cpu.prof"
echo
echo "To view memory profile:"
echo "  go tool pprof -http=:8080 ${RESULTS_DIR}/ticket_manager_mem.prof"