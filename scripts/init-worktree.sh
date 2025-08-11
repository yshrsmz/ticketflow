#!/usr/bin/env bash

# init-worktree.sh - Initialize a new git worktree with necessary setup
# This script is called after a new worktree is created by ticketflow

set -euo pipefail

# Current working directory should be the worktree directory
WORKTREE_DIR="$(pwd)"

# Get the main repository directory using git worktree
MAIN_REPO_DIR="$(git worktree list --porcelain | grep -E '^worktree ' | head -1 | cut -d' ' -f2-)"

# Fallback if git worktree command fails
if [ -z "$MAIN_REPO_DIR" ] || [ ! -d "$MAIN_REPO_DIR" ]; then
    echo "Error: Unable to determine main repository directory"
    exit 1
fi

# Ensure we're not running in the main repository
if [ "$WORKTREE_DIR" = "$MAIN_REPO_DIR" ]; then
    echo "Error: This script should not be run in the main repository"
    echo "Please run it from within a git worktree"
    exit 1
fi

echo "Initializing worktree at: $WORKTREE_DIR"

# Define files to symlink from main repository
# Add more files here as needed
SYMLINK_TARGETS=(
    ".claude/settings.local.json"
    ".env"
)

# Create symlinks for all target files
for TARGET in "${SYMLINK_TARGETS[@]}"; do
    # Get directory part of the target path
    TARGET_DIR="$(dirname "$TARGET")"

    # Create directory if needed and it doesn't exist
    if [ "$TARGET_DIR" != "." ] && [ ! -d "$WORKTREE_DIR/$TARGET_DIR" ]; then
        mkdir -p "$WORKTREE_DIR/$TARGET_DIR"
    fi

    MAIN_FILE="$MAIN_REPO_DIR/$TARGET"
    WORKTREE_FILE="$WORKTREE_DIR/$TARGET"

    if [ -f "$MAIN_FILE" ]; then
        if [ ! -e "$WORKTREE_FILE" ]; then
            ln -s "$MAIN_FILE" "$WORKTREE_FILE"
            echo "Created symlink: $TARGET -> $MAIN_FILE"
        else
            echo "$TARGET already exists, skipping symlink creation"
        fi
    else
        echo "Warning: $MAIN_FILE not found, skipping symlink creation"
    fi
done

echo "Worktree initialization complete"
