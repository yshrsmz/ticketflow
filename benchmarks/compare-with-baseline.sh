#!/bin/bash

# compare-with-baseline.sh - Compare current performance with baseline
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

BASELINE_FILE="benchmarks/baseline.txt"
CURRENT_FILE="benchmarks/current.txt"
THRESHOLD_TIME=10  # 10% regression threshold for time
THRESHOLD_ALLOC=20 # 20% regression threshold for allocations

# Check if baseline exists
if [ ! -f "${BASELINE_FILE}" ]; then
    echo -e "${RED}Error: No baseline file found at ${BASELINE_FILE}${NC}"
    echo "Run 'make bench-baseline' to create a baseline first."
    exit 1
fi

echo -e "${GREEN}Comparing current performance with baseline...${NC}"
echo

# Run current benchmarks
echo -e "${YELLOW}Running current benchmarks...${NC}"
go test -bench=. -benchmem -benchtime=3s -run=^$ ./internal/cli ./internal/ticket ./internal/git > "${CURRENT_FILE}" 2>&1

# Function to extract benchmark value
extract_value() {
    local line=$1
    local field=$2
    echo "$line" | awk "{print \$${field}}"
}

# Function to calculate percentage change
calc_percentage() {
    local old=$1
    local new=$2
    if [ "$old" == "0" ]; then
        echo "0"
    else
        echo "scale=2; ((${new} - ${old}) / ${old}) * 100" | bc
    fi
}

# Compare benchmarks
echo -e "${YELLOW}Analyzing results...${NC}"
echo

regression_found=false
improvements_found=false

# Process each benchmark in current results
grep "^Benchmark" "${CURRENT_FILE}" | while IFS= read -r current_line; do
    bench_name=$(extract_value "$current_line" 1)
    
    # Find corresponding baseline
    baseline_line=$(grep "^${bench_name}" "${BASELINE_FILE}" 2>/dev/null || true)
    
    if [ -z "${baseline_line}" ]; then
        echo -e "${YELLOW}NEW:${NC} ${bench_name} (no baseline)"
        continue
    fi
    
    # Extract metrics
    current_ns=$(extract_value "$current_line" 3)
    baseline_ns=$(extract_value "$baseline_line" 3)
    
    # Remove units (ns/op) if present
    current_ns=${current_ns%ns/op}
    baseline_ns=${baseline_ns%ns/op}
    
    # Calculate percentage change
    if [[ "$current_ns" =~ ^[0-9]+\.?[0-9]*$ ]] && [[ "$baseline_ns" =~ ^[0-9]+\.?[0-9]*$ ]]; then
        change=$(calc_percentage "$baseline_ns" "$current_ns")
        
        # Format output based on change
        if (( $(echo "$change > ${THRESHOLD_TIME}" | bc -l) )); then
            echo -e "${RED}REGRESSION:${NC} ${bench_name}"
            echo "  Time: ${baseline_ns} ns/op → ${current_ns} ns/op (+${change}%)"
            regression_found=true
        elif (( $(echo "$change < -5" | bc -l) )); then
            echo -e "${GREEN}IMPROVED:${NC} ${bench_name}"
            echo "  Time: ${baseline_ns} ns/op → ${current_ns} ns/op (${change}%)"
            improvements_found=true
        fi
    fi
    
    # Check allocations if present
    if [[ "$current_line" == *"allocs/op"* ]] && [[ "$baseline_line" == *"allocs/op"* ]]; then
        current_allocs=$(echo "$current_line" | grep -oE '[0-9]+ allocs/op' | awk '{print $1}')
        baseline_allocs=$(echo "$baseline_line" | grep -oE '[0-9]+ allocs/op' | awk '{print $1}')
        
        if [ -n "$current_allocs" ] && [ -n "$baseline_allocs" ]; then
            alloc_change=$(calc_percentage "$baseline_allocs" "$current_allocs")
            
            if (( $(echo "$alloc_change > ${THRESHOLD_ALLOC}" | bc -l) )); then
                echo "  ${RED}Allocations: ${baseline_allocs} → ${current_allocs} (+${alloc_change}%)${NC}"
                regression_found=true
            elif (( $(echo "$alloc_change < -10" | bc -l) )); then
                echo "  ${GREEN}Allocations: ${baseline_allocs} → ${current_allocs} (${alloc_change}%)${NC}"
            fi
        fi
    fi
done

# Summary
echo
echo -e "${YELLOW}=== Summary ===${NC}"

if [ "$regression_found" = true ]; then
    echo -e "${RED}Performance regressions detected!${NC}"
    echo "Some benchmarks are slower than the baseline by more than ${THRESHOLD_TIME}%"
    exit_code=1
else
    echo -e "${GREEN}No significant regressions detected.${NC}"
    exit_code=0
fi

if [ "$improvements_found" = true ]; then
    echo -e "${GREEN}Performance improvements found!${NC}"
    echo "Consider updating the baseline with: make bench-baseline"
fi

echo
echo "Full results saved to: ${CURRENT_FILE}"
echo "Baseline file: ${BASELINE_FILE}"

exit $exit_code