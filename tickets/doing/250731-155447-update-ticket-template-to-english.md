---
priority: 2
description: Update the default ticket template in the code to use English instead of Japanese
created_at: "2025-07-31T15:54:47+09:00"
started_at: "2025-07-31T15:58:52+09:00"
closed_at: null
---

# Ticket Overview

The default ticket template in `internal/config/config.go` contains Japanese text. While the local `.ticketflow.yaml` has an English template, the hardcoded default in the Go code should also be in English for consistency and broader accessibility.

## Current Situation

1. **Default template in code** (`internal/config/config.go` lines 64-79):
   - Contains Japanese headers: 概要 (Summary), タスク (Tasks), 技術仕様 (Technical Specifications), メモ (Notes)
   - This is used when no local config exists

2. **Local config template** (`.ticketflow.yaml`):
   - Already in English
   - Has more comprehensive task checklist

## Tasks
- [ ] Update the default template in `internal/config/config.go` to use English headers
- [ ] Match the structure with the local config template for consistency
- [ ] Run `make test` to run the tests
- [ ] Run `make vet`, `make fmt` and `make lint`
- [ ] Verify the change works correctly when no local config exists
- [ ] Get developer approval before closing

## Implementation Plan

Replace the Japanese template in `internal/config/config.go` with:
```
# Summary

[Describe the ticket summary here]

## Tasks
- [ ] Task 1
- [ ] Task 2
- [ ] Task 3

## Technical Specifications

[Add technical details as needed]

## Notes

[Additional notes or remarks]
```

## Notes

This change improves international accessibility while maintaining the same structure and functionality.