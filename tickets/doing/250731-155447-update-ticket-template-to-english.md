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
- [x] Update the default template in `internal/config/config.go` to use English headers
- [x] Match the structure with the local config template for consistency
- [x] Run `make test` to run the tests
- [x] Run `make vet`, `make fmt` and `make lint`
- [x] Verify the change works correctly when no local config exists
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

## Implementation Details

The change was successfully implemented:
- Updated `internal/config/config.go` lines 64-79 to replace Japanese headers with English
- All tests pass successfully
- Code quality checks (vet, fmt, lint) all pass
- Verified the English template is used when no local config exists

The commit has been made with message: "Update default ticket template to English"