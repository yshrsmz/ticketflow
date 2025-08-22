#!/bin/bash

# setup-lefthook.sh - Install and configure Lefthook for the ticketflow project
# This script primarily uses Homebrew for installation with fallback options

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
LEFTHOOK_VERSION="1.12.3"
GITHUB_REPO="evilmartians/lefthook"

# Helper functions
print_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Detect OS
detect_os() {
    case "$(uname -s)" in
        Darwin*)    echo "darwin" ;;
        Linux*)     echo "linux" ;;
        CYGWIN*|MINGW*|MSYS*) echo "windows" ;;
        *)          echo "unknown" ;;
    esac
}

# Detect architecture
detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)   echo "amd64" ;;
        arm64|aarch64)  echo "arm64" ;;
        *)              echo "unknown" ;;
    esac
}

# Check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if Lefthook is already installed
check_existing_installation() {
    if command_exists lefthook; then
        local current_version=$(lefthook version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
        print_info "Lefthook is already installed (version: $current_version)"
        
        read -p "Do you want to reinstall/update? (y/N): " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Yy]$ ]]; then
            print_info "Keeping existing installation"
            return 0
        fi
    fi
    return 1
}

# Install using Go
install_with_go() {
    if command_exists go; then
        print_info "Installing Lefthook using go install..."
        if go install github.com/evilmartians/lefthook@v${LEFTHOOK_VERSION}; then
            print_success "Lefthook installed via go install"
            return 0
        else
            print_warning "Failed to install via go install, trying next method..."
        fi
    else
        print_info "Go not found, skipping go install method"
    fi
    return 1
}

# Install using Homebrew (macOS and Linux)
install_with_homebrew() {
    if command_exists brew; then
        print_info "Installing Lefthook using Homebrew..."
        if brew install lefthook; then
            print_success "Lefthook installed via Homebrew"
            return 0
        else
            print_warning "Failed to install via Homebrew, trying next method..."
        fi
    else
        if [[ "$1" == "darwin" ]]; then
            print_warning "Homebrew not found. Install it from https://brew.sh"
            print_info "Run: /bin/bash -c \"\$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)\""
        else
            print_info "Homebrew not found, skipping Homebrew method"
        fi
    fi
    return 1
}

# Download binary directly from GitHub
install_binary() {
    local os=$1
    local arch=$2
    
    print_info "Downloading Lefthook binary from GitHub..."
    
    # Construct download URL based on OS and architecture
    local binary_name="lefthook"
    local download_url=""
    
    # Try to construct the URL based on common patterns
    if [[ "$os" == "darwin" ]]; then
        # For macOS, binaries might be named differently
        download_url="https://github.com/${GITHUB_REPO}/releases/download/v${LEFTHOOK_VERSION}/lefthook_${LEFTHOOK_VERSION}_${os}_${arch}"
    elif [[ "$os" == "linux" ]]; then
        download_url="https://github.com/${GITHUB_REPO}/releases/download/v${LEFTHOOK_VERSION}/lefthook_${LEFTHOOK_VERSION}_${os}_${arch}"
    else
        print_error "Unsupported OS for binary download: $os"
        return 1
    fi
    
    # Create temporary directory
    local tmp_dir=$(mktemp -d)
    trap "rm -rf $tmp_dir" EXIT
    
    # Download binary
    print_info "Downloading from: $download_url"
    if curl -L -f -o "$tmp_dir/lefthook" "$download_url" 2>/dev/null || \
       curl -L -f -o "$tmp_dir/lefthook" "${download_url}.tar.gz" 2>/dev/null || \
       curl -L -f -o "$tmp_dir/lefthook" "${download_url}.gz" 2>/dev/null; then
        
        # Handle compressed files
        if [[ -f "$tmp_dir/lefthook" ]]; then
            local file_type=$(file "$tmp_dir/lefthook" | cut -d: -f2)
            if echo "$file_type" | grep -q "gzip"; then
                print_info "Extracting compressed binary..."
                gunzip "$tmp_dir/lefthook"
                mv "$tmp_dir/lefthook" "$tmp_dir/lefthook.gz"
                gunzip "$tmp_dir/lefthook.gz"
            elif echo "$file_type" | grep -q "tar"; then
                print_info "Extracting tar archive..."
                tar -xzf "$tmp_dir/lefthook" -C "$tmp_dir"
                rm "$tmp_dir/lefthook"
                # Find the extracted binary
                find "$tmp_dir" -name "lefthook" -type f -exec mv {} "$tmp_dir/lefthook" \;
            fi
        fi
        
        # Make binary executable
        chmod +x "$tmp_dir/lefthook"
        
        # Move to appropriate location
        local install_dir=""
        if [[ -w "/usr/local/bin" ]]; then
            install_dir="/usr/local/bin"
        elif [[ -d "$HOME/.local/bin" ]]; then
            install_dir="$HOME/.local/bin"
        elif [[ -d "$HOME/bin" ]]; then
            install_dir="$HOME/bin"
        else
            print_info "Creating $HOME/.local/bin for installation"
            mkdir -p "$HOME/.local/bin"
            install_dir="$HOME/.local/bin"
        fi
        
        print_info "Installing to $install_dir/lefthook"
        if [[ -w "$install_dir" ]]; then
            mv "$tmp_dir/lefthook" "$install_dir/lefthook"
        else
            print_warning "Need sudo access to install to $install_dir"
            sudo mv "$tmp_dir/lefthook" "$install_dir/lefthook"
        fi
        
        # Add to PATH if needed
        if ! echo "$PATH" | grep -q "$install_dir"; then
            print_warning "$install_dir is not in PATH"
            print_info "Add the following line to your shell configuration:"
            print_info "  export PATH=\"$install_dir:\$PATH\""
        fi
        
        print_success "Lefthook binary installed successfully"
        return 0
    else
        print_error "Failed to download Lefthook binary"
        return 1
    fi
}

# Main installation flow
main() {
    print_info "Setting up Lefthook for ticketflow project"
    
    # Check existing installation
    if check_existing_installation; then
        # User chose to keep existing installation
        if command_exists lefthook; then
            print_info "Installing git hooks..."
            lefthook install
            print_success "Git hooks installed successfully"
            return 0
        fi
    fi
    
    # Detect OS and architecture
    local os=$(detect_os)
    local arch=$(detect_arch)
    
    print_info "Detected OS: $os, Architecture: $arch"
    
    if [[ "$os" == "unknown" ]] || [[ "$arch" == "unknown" ]]; then
        print_error "Unable to detect OS or architecture"
        exit 1
    fi
    
    # Try installation methods in order of preference
    # 1. Homebrew (preferred for macOS and Linux)
    if install_with_homebrew "$os"; then
        :  # Successfully installed
    # 2. Go install (for Go developers)
    elif install_with_go; then
        :  # Successfully installed
    # 3. Direct binary download (last resort)
    elif install_binary "$os" "$arch"; then
        :  # Successfully installed
    else
        print_error "Failed to install Lefthook using all available methods"
        print_info "Please install Lefthook manually:"
        print_info "  - Using Homebrew (recommended): brew install lefthook"
        print_info "  - Using Go: go install github.com/evilmartians/lefthook@latest"
        print_info "  - Download from: https://github.com/evilmartians/lefthook/releases"
        exit 1
    fi
    
    # Verify installation
    if command_exists lefthook; then
        print_success "Lefthook installed successfully!"
        print_info "Version: $(lefthook version)"
        
        # Install git hooks
        print_info "Installing git hooks..."
        lefthook install
        print_success "Git hooks installed successfully"
        
        print_info ""
        print_success "Setup complete! Lefthook is now managing your git hooks."
        print_info "Configuration file: lefthook.yml"
        print_info "To skip hooks temporarily, use: git commit --no-verify"
    else
        print_error "Lefthook installation could not be verified"
        exit 1
    fi
}

# Run main function
main "$@"