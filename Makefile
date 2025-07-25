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
VERSION?=dev
LDFLAGS=-ldflags "-X main.Version=$(VERSION)"

.PHONY: all build test clean install run run-tui

# Default target
all: test build

# Build the binary
build:
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_NAME) $(MAIN_PACKAGE)

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

# Build for multiple platforms
build-all:
	GOOS=darwin GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-amd64 $(MAIN_PACKAGE)
	GOOS=darwin GOARCH=arm64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-darwin-arm64 $(MAIN_PACKAGE)
	GOOS=linux GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	GOOS=windows GOARCH=amd64 $(GOBUILD) $(LDFLAGS) -o dist/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)

# Show help
help:
	@echo "Available targets:"
	@echo "  make build     - Build the binary"
	@echo "  make install   - Install the binary to GOPATH/bin"
	@echo "  make test      - Run all tests"
	@echo "  make coverage  - Generate test coverage report"
	@echo "  make run       - Build and run the application"
	@echo "  make clean     - Remove build artifacts"
	@echo "  make deps      - Download dependencies"
	@echo "  make fmt       - Format Go code"
	@echo "  make lint      - Run linter"