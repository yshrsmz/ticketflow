#!/bin/bash

# init-githooks.sh - Install native git hooks for ticketflow project
# Creates symlinks from .git/hooks to scripts/githooks

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Helper functions
print_info() {
    echo -e "${BLUE}[githooks]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[githooks]${NC} ✓ $1"
}

print_warning() {
    echo -e "${YELLOW}[githooks]${NC} ⚠ $1"
}

print_error() {
    echo -e "${RED}[githooks]${NC} ✗ $1"
}

# Get the script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
HOOKS_SOURCE_DIR="$PROJECT_ROOT/scripts/githooks"

# Determine the git hooks directory
find_git_hooks_dir() {
    local git_dir
    
    # First check if we're in a worktree
    if [[ -f ".git" ]]; then
        # This is a worktree, .git is a file pointing to the real git dir
        git_dir=$(cat .git | sed 's/gitdir: //')
        # Handle relative paths
        if [[ ! "$git_dir" = /* ]]; then
            git_dir="$PROJECT_ROOT/$git_dir"
        fi
        echo "$git_dir/hooks"
    elif [[ -d ".git" ]]; then
        # Regular repository
        echo "$PROJECT_ROOT/.git/hooks"
    else
        print_error "Not in a git repository"
        exit 1
    fi
}

# Parse command line arguments
ACTION="install"
if [[ "$1" == "uninstall" ]] || [[ "$1" == "remove" ]]; then
    ACTION="uninstall"
elif [[ "$1" == "help" ]] || [[ "$1" == "--help" ]] || [[ "$1" == "-h" ]]; then
    echo "Usage: $0 [install|uninstall|help]"
    echo ""
    echo "Commands:"
    echo "  install    Install git hooks (default)"
    echo "  uninstall  Remove git hooks"
    echo "  help       Show this help message"
    echo ""
    echo "Environment variables:"
    echo "  HOOKS_VERBOSE=1  Enable verbose output in hooks"
    echo "  SKIP_HOOKS=1     Skip hook execution"
    echo "  NO_VERIFY=1      Skip hook execution (git standard)"
    exit 0
fi

# Change to project root
cd "$PROJECT_ROOT"

# Find the git hooks directory
HOOKS_DIR=$(find_git_hooks_dir)
print_info "Git hooks directory: $HOOKS_DIR"

# Create hooks directory if it doesn't exist
if [[ ! -d "$HOOKS_DIR" ]]; then
    print_info "Creating hooks directory..."
    mkdir -p "$HOOKS_DIR"
fi

# List of hooks to manage
HOOKS=(
    "pre-commit"
    "pre-push"
)

# Uninstall function
uninstall_hooks() {
    print_info "Uninstalling git hooks..."
    
    for hook in "${HOOKS[@]}"; do
        hook_path="$HOOKS_DIR/$hook"
        
        if [[ -L "$hook_path" ]]; then
            # It's a symlink, check if it points to our hook
            target=$(readlink "$hook_path")
            if [[ "$target" == *"scripts/githooks/$hook" ]]; then
                rm "$hook_path"
                print_success "Removed $hook"
            else
                print_warning "$hook points elsewhere, skipping: $target"
            fi
        elif [[ -f "$hook_path" ]]; then
            print_warning "$hook exists but is not a symlink, skipping"
        else
            print_info "$hook not installed"
        fi
    done
    
    print_success "Git hooks uninstalled"
}

# Install function
install_hooks() {
    print_info "Installing git hooks..."
    
    # Check if source hooks exist
    if [[ ! -d "$HOOKS_SOURCE_DIR" ]]; then
        print_error "Hooks source directory not found: $HOOKS_SOURCE_DIR"
        exit 1
    fi
    
    for hook in "${HOOKS[@]}"; do
        source_path="$HOOKS_SOURCE_DIR/$hook"
        hook_path="$HOOKS_DIR/$hook"
        
        if [[ ! -f "$source_path" ]]; then
            print_warning "Source hook not found: $source_path"
            continue
        fi
        
        # Make source hook executable
        chmod +x "$source_path"
        
        # Check if hook already exists
        if [[ -L "$hook_path" ]]; then
            # It's a symlink
            target=$(readlink "$hook_path")
            if [[ "$target" == "$source_path" ]]; then
                print_info "$hook already installed correctly"
                continue
            else
                print_warning "$hook exists but points elsewhere: $target"
                read -p "Replace with our hook? (y/N): " -n 1 -r
                echo
                if [[ $REPLY =~ ^[Yy]$ ]]; then
                    rm "$hook_path"
                else
                    print_warning "Skipping $hook"
                    continue
                fi
            fi
        elif [[ -f "$hook_path" ]]; then
            # Regular file exists
            print_warning "$hook already exists (not a symlink)"
            
            # Show first few lines of existing hook
            echo "Existing hook preview:"
            head -n 5 "$hook_path" | sed 's/^/  /'
            
            read -p "Replace with our hook? (y/N): " -n 1 -r
            echo
            if [[ $REPLY =~ ^[Yy]$ ]]; then
                # Backup existing hook
                backup_path="${hook_path}.backup.$(date +%Y%m%d_%H%M%S)"
                mv "$hook_path" "$backup_path"
                print_info "Backed up existing hook to: $(basename "$backup_path")"
            else
                print_warning "Skipping $hook"
                continue
            fi
        fi
        
        # Create symlink
        ln -s "$source_path" "$hook_path"
        print_success "Installed $hook"
    done
    
    print_success "Git hooks installed successfully!"
    print_info ""
    print_info "Hook commands:"
    print_info "  • Skip hooks once:  git commit --no-verify"
    print_info "  • Skip all hooks:   export SKIP_HOOKS=1"
    print_info "  • Verbose output:   export HOOKS_VERBOSE=1"
    print_info "  • Uninstall hooks:  $0 uninstall"
}

# Main execution
if [[ "$ACTION" == "uninstall" ]]; then
    uninstall_hooks
else
    install_hooks
fi