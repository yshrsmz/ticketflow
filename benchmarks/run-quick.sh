#!/bin/bash

# run-quick.sh - Run quick benchmarks for rapid feedback
set -euo pipefail

# Check for required commands
command -v go >/dev/null 2>&1 || { echo "Error: 'go' command not found. Please install Go."; exit 1; }

# Colors for output (with color support detection)
if [ -t 1 ] && [ "${TERM:-}" != "dumb" ] && command -v tput >/dev/null 2>&1 && [ "$(tput colors 2>/dev/null || echo 0)" -ge 8 ]; then
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    NC='\033[0m' # No Color
else
    GREEN=''
    YELLOW=''
    NC=''
fi

# Allow BENCH_TIME to be configured via environment variable (default: 1s)
BENCH_TIME="${BENCH_TIME:-1s}"

echo -e "${GREEN}Running quick benchmarks (BENCH_TIME=${BENCH_TIME})...${NC}"
echo

# Quick benchmarks with shorter runtime
PACKAGES=(
    "./internal/cli"
    "./internal/ticket"
    "./internal/git"
)

for pkg in "${PACKAGES[@]}"; do
    echo -e "${YELLOW}Benchmarking ${pkg}...${NC}"
    go test -bench=. -benchmem -benchtime="${BENCH_TIME}" -run=^$ "${pkg}" | grep -E "^Benchmark|^ok|^PASS|^FAIL" || true
    echo
done

echo -e "${GREEN}Quick benchmarks complete!${NC}"