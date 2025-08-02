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

# Create .claude directory if it doesn't exist
if [ ! -d "$WORKTREE_DIR/.claude" ]; then
    mkdir -p "$WORKTREE_DIR/.claude"
fi

# Create symlink for .claude/settings.local.json
MAIN_SETTINGS="$MAIN_REPO_DIR/.claude/settings.local.json"
WORKTREE_SETTINGS="$WORKTREE_DIR/.claude/settings.local.json"

if [ -f "$MAIN_SETTINGS" ]; then
    if [ ! -e "$WORKTREE_SETTINGS" ]; then
        ln -s "$MAIN_SETTINGS" "$WORKTREE_SETTINGS"
        echo "Created symlink: .claude/settings.local.json -> $MAIN_SETTINGS"
    else
        echo ".claude/settings.local.json already exists, skipping symlink creation"
    fi
else
    echo "Warning: $MAIN_SETTINGS not found, skipping symlink creation"
fi

echo "Worktree initialization complete"