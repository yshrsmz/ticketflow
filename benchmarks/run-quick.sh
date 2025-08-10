#!/bin/bash

# run-quick.sh - Run quick benchmarks for rapid feedback
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Running quick benchmarks...${NC}"
echo

# Quick benchmarks with shorter runtime
PACKAGES=(
    "./internal/cli"
    "./internal/ticket"
    "./internal/git"
)

for pkg in "${PACKAGES[@]}"; do
    echo -e "${YELLOW}Benchmarking ${pkg}...${NC}"
    go test -bench=. -benchmem -benchtime=1s -run=^$ "${pkg}" | grep -E "^Benchmark|^ok|^PASS|^FAIL" || true
    echo
done

echo -e "${GREEN}Quick benchmarks complete!${NC}"