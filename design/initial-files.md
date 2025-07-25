# TicketFlow Initial Files

## 1. .ticketflow.yaml.example

```yaml
# TicketFlow Configuration File
# Copy this file to .ticketflow.yaml and customize as needed

# Git settings
git:
  # Default branch to merge tickets into
  default_branch: "main"
  
# Worktree settings
worktree:
  # Enable git worktree integration
  enabled: true
  
  # Base directory for worktrees (relative or absolute path)
  base_dir: "../.worktrees"
  
  # Commands to run after creating a new worktree
  # These run in the worktree directory
  init_commands:
    - "git pull origin main --rebase"
    # - "npm install"
    # - "make deps"
    
  # Automatic operations
  auto_operations:
    # Create worktree when starting a ticket
    create_on_start: true
    
    # Remove worktree when closing a ticket
    remove_on_close: true
    
    # Clean up orphaned worktrees periodically
    cleanup_orphaned: true
    
# Ticket settings  
tickets:
  # Directory for ticket files
  dir: "tickets"
  
  # Directory for archived (done) tickets
  archive_dir: "tickets/done"
  
  # Default content for new tickets
  template: |
    # 概要
    
    [ここにチケットの概要を記述]
    
    ## タスク
    - [ ] タスク1
    - [ ] タスク2
    - [ ] タスク3
    
    ## 技術仕様
    
    [必要に応じて技術的な詳細を記述]
    
    ## メモ
    
    [追加の注意事項やメモ]

# Output settings
output:
  # Default output format for list/show commands
  default_format: "text"  # text or json
  
  # Pretty print JSON output
  json_pretty: true
```

## 2. README.md

```markdown
# TicketFlow

A git worktree-based ticket management system optimized for AI collaboration.

## Features

- 📝 Markdown-based tickets with YAML frontmatter
- 🌳 Git worktree integration for parallel work
- 🤖 AI-friendly CLI with JSON output
- 🎨 Beautiful TUI for human interaction
- 🔄 Seamless workflow from creation to completion

## Installation

```bash
go install github.com/yshrsmz/ticketflow/cmd/ticketflow@latest
```

Or build from source:

```bash
git clone https://github.com/yshrsmz/ticketflow.git
cd ticketflow
make install
```

## Quick Start

1. Initialize in your project:
   ```bash
   cd your-project
   ticketflow init
   ```

2. Create a ticket:
   ```bash
   ticketflow new implement-feature
   ```

3. Start working:
   ```bash
   ticketflow start 250124-150000-implement-feature
   ```

4. Complete work:
   ```bash
   ticketflow close
   ```

## TUI Mode

Run without arguments to start the interactive TUI:

```bash
ticketflow
```

### Keyboard Shortcuts

- `n` - New ticket
- `s` - Start work on selected ticket
- `Enter` - View ticket details
- `w` - Show worktrees
- `/` - Search tickets
- `?` - Help
- `q` - Quit

## CLI Commands

### Basic Commands

```bash
ticketflow init                     # Initialize system
ticketflow new <slug>               # Create ticket
ticketflow list                     # List tickets
ticketflow start <ticket-id>        # Start work
ticketflow close                    # Complete work
ticketflow restore                  # Restore symlink
```

### Advanced Options

```bash
# List with filters
ticketflow list --status doing --format json

# Start without pushing
ticketflow start 250124-150000-feature --no-push

# Force close with uncommitted changes
ticketflow close --force
```

## AI Integration

TicketFlow is designed for AI agents:

```bash
# Get JSON output for parsing
ticketflow list --format json

# Check current status
ticketflow status --format json

# AI-friendly error messages
{
  "error": {
    "code": "TICKET_NOT_FOUND",
    "message": "Ticket not found",
    "suggestions": ["Check ticket ID", "Run 'ticketflow list'"]
  }
}
```

## Configuration

Edit `.ticketflow.yaml`:

```yaml
git:
  default_branch: "main"
  
worktree:
  enabled: true
  base_dir: "../.worktrees"
  init_commands:
    - "npm install"
    
tickets:
  dir: "tickets"
  template: |
    # Your custom template
```

## Development

```bash
# Run tests
make test

# Build
make build

# Run TUI in development
make run-tui
```

## License

MIT
```

## 3. go.mod (初期状態)

```go
module github.com/yshrsmz/ticketflow

go 1.21

require (
    github.com/charmbracelet/bubbletea v0.25.0
    github.com/charmbracelet/lipgloss v0.10.0
    github.com/stretchr/testify v1.9.0
    gopkg.in/yaml.v3 v3.0.1
)
```

## 4. Makefile

```makefile
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
```

## 5. 実装開始時のディレクトリ作成スクリプト

```bash
#!/bin/bash
# setup.sh - Initial project setup

# Create directory structure
mkdir -p cmd/ticketflow
mkdir -p internal/{config,ticket,git,ui/views,ui/styles,ui/components,cli}
mkdir -p test/{integration,e2e,testutil}

# Create .gitignore
cat > .gitignore << 'EOF'
# Binaries
ticketflow
*.exe
*.dll
*.so
*.dylib

# Test binary
*.test

# Coverage
*.out
coverage.html

# IDE
.idea/
.vscode/
*.swp

# OS
.DS_Store
Thumbs.db

# Project specific
current-ticket.md
.worktrees/
*.log
dist/
.ticketflow.local.yaml
EOF

# Initialize go module
go mod init github.com/yshrsmz/ticketflow

# Add dependencies
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/stretchr/testify
go get gopkg.in/yaml.v3

# Create example config
cp .ticketflow.yaml.example .ticketflow.yaml.example

echo "Project structure created successfully!"
echo "Next steps:"
echo "1. Update go.mod with your actual module path"
echo "2. Start implementing from Phase 1 (core functionality)"
```

---

これらのドキュメントをClaude Codeで参照しながら実装を進めてください。実装の順序は：

1. **Phase 1**: コア機能（config, ticket model, basic CLI）
2. **Phase 2**: Worktree統合
3. **Phase 3**: TUI実装
4. **Phase 4**: 高度な機能

各フェーズが完了したらテストを実行し、動作確認をしてから次に進んでください。頑張ってください！
```
