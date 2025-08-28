# Project variables
BINARY_NAME=ticketflow
GO_FILES=$(shell find . -name '*.go' -type f)
MAIN_PACKAGE=./cmd/ticketflow

# Go commands
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Build variables
# Get version from git tags or use dev
VERSION := $(shell git describe --tags --always --dirty=-dev 2>/dev/null || echo "dev")
BUILD_TIME := $(shell date -u '+%Y-%m-%d %H:%M:%S')
GIT_COMMIT := $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS=-ldflags "-X main.Version=$(VERSION) -X 'main.BuildTime=$(BUILD_TIME)' -X main.GitCommit=$(GIT_COMMIT)"

# Platform detection
CURRENT_OS := $(shell go env GOOS)
CURRENT_ARCH := $(shell go env GOARCH)

# Benchmark variables
BENCH_TIME ?= 3s
BENCH_COUNT ?= 3
BENCH_TIMEOUT ?= 30m

.PHONY: all build test clean install run run-tui build-current build-linux build-mac build-all release-archives init-worktree setup-hooks init

# Default target
all: test build

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

# Build for current platform with architecture in filename
build-current:
	@echo "Building for $(CURRENT_OS)/$(CURRENT_ARCH)..."
	@mkdir -p dist
	$(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-$(CURRENT_OS)-$(CURRENT_ARCH) $(MAIN_PACKAGE)
	@echo "Built: dist/$(BINARY_NAME)-$(CURRENT_OS)-$(CURRENT_ARCH)"

# Install the binary
install: build
	mkdir -p $(GOPATH)/bin
	cp $(BINARY_NAME) $(GOPATH)/bin/

# Run tests
test:
	$(GOTEST) -v ./...

# Run unit tests only
test-unit:
	$(GOTEST) -v -race ./internal/...

# Run integration tests
test-integration:
	$(GOTEST) -v -race ./test/integration/...

# Run E2E tests
test-e2e: build
	$(GOTEST) -v ./test/e2e/...

# Generate test coverage
coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Run simple benchmark (manual use only)
bench:
	$(GOTEST) -bench=BenchmarkManagerList -benchmem -run=^$$ ./internal/ticket

# Run go fmt
fmt:
	$(GOFMT) ./...

# Run go vet
vet:
	$(GOVET) ./...

# Run linter (requires golangci-lint)
lint:
	golangci-lint run

# Clean build artifacts
clean:
	rm -f $(BINARY_NAME)
	rm -f coverage.out coverage.html
	rm -rf dist/

# Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

# Run the application
run: build
	./$(BINARY_NAME)

# Run TUI mode
run-tui: build
	./$(BINARY_NAME)

# Run CLI with arguments
run-cli: build
	./$(BINARY_NAME) $(ARGS)

# Development watch mode (requires entr)
watch:
	find . -name '*.go' | entr -r make run

# Initialize worktree (symlink Claude settings)
init-worktree:
	@bash scripts/init-worktree.sh

# Setup native git hooks
setup-hooks:
	@echo "Setting up native git hooks..."
	@bash scripts/init-githooks.sh
	@echo "Git hooks setup complete!"

# Initialize development environment (dependencies + hooks + worktree)
init: deps setup-hooks init-worktree
	@echo "Development environment initialized successfully!"

# Build for Linux platforms
build-linux:
	@echo "Building for Linux platforms..."
	@mkdir -p dist
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	@echo "Linux builds complete"

# Build for macOS platforms
build-mac:
	@echo "Building for macOS platforms..."
	@mkdir -p dist
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	@echo "macOS builds complete"

# Build for multiple platforms
build-all: build-linux build-mac
	@echo "Building for Windows..."
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "All platform builds complete"

# Show version
version:
	@echo "Version: $(VERSION)"
	@echo "Git Commit: $(GIT_COMMIT)"
	@echo "Build Time: $(BUILD_TIME)"

# Create a new release tag
release:
	@if [ -z "$(TAG)" ]; then echo "Usage: make release TAG=v0.1.0"; exit 1; fi
	@echo "Creating release $(TAG)..."
	@git tag -a $(TAG) -m "Release $(TAG)"
	@echo "Release $(TAG) created. Push with: git push origin $(TAG)"

# Build release binaries
release-build: clean
	@mkdir -p dist
	@echo "Building release binaries for version $(VERSION)..."
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-linux-arm64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-$(VERSION)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "Release binaries built in dist/"

# Create compressed archives for releases
release-archives: release-build
	@echo "Creating release archives..."
	@cd dist && \
	for file in $(BINARY_NAME)-$(VERSION)-*; do \
		if [ "$${file##*.}" = "exe" ]; then \
			echo "Creating zip for $$file..."; \
			zip -q "$${file%.exe}.zip" "$$file"; \
			rm "$$file"; \
		else \
			echo "Creating tar.gz for $$file..."; \
			tar -czf "$$file.tar.gz" "$$file"; \
			rm "$$file"; \
		fi; \
	done
	@echo "Generating checksums..."
	@cd dist && shasum -a 256 *.tar.gz *.zip > checksums.txt 2>/dev/null || sha256sum *.tar.gz *.zip > checksums.txt
	@echo "Release archives created in dist/"
	@echo "Contents:"
	@ls -la dist/*.tar.gz dist/*.zip 2>/dev/null || true
	@echo ""
	@echo "Checksums:"
	@cat dist/checksums.txt

# Show help
help:
	@echo "Available targets:"
	@echo "  make init          - Initialize development environment (deps, hooks, worktree)"
	@echo "  make setup-hooks   - Setup native git hooks"
	@echo "  make build         - Build the binary for local development"
	@echo "  make build-current - Build for current platform with arch in filename"
	@echo "  make build-linux   - Build for Linux (amd64, arm64)"
	@echo "  make build-mac     - Build for macOS (amd64, arm64)"
	@echo "  make build-all     - Build for all platforms"
	@echo "  make install       - Install the binary to GOPATH/bin"
	@echo "  make test          - Run all tests"
	@echo "  make coverage      - Generate test coverage report"
	@echo "  make bench         - Run simple benchmark (manual use)"
	@echo "  make run           - Build and run the application"
	@echo "  make clean         - Remove build artifacts"
	@echo "  make deps          - Download dependencies"
	@echo "  make fmt           - Format Go code"
	@echo "  make vet           - Run go vet"
	@echo "  make lint          - Run linter"
	@echo "  make init-worktree - Initialize worktree with Claude settings"
	@echo "  make version       - Show version information"
	@echo "  make release       - Create a new release tag (TAG=v0.1.0)"
	@echo "  make release-build - Build release binaries for all platforms"
	@echo "  make release-archives - Build and create compressed archives"
