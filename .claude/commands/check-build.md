---
description: "check build, test, linter"
---

- Ensure debug build succeeds: !"make build"
  - If it fails, fix until build succeeds. Even if it looks unrelated, you must fix.
- Ensure all test succeeds: !"make test"
  - If it fails, fix until test succeeds. Even if it looks unrelated, you must fix.
- Ensure lint task succeeds: !"make lint"
  - If it fails, fix until lint succeeds. Even if it looks unrelated, you must fix.
- Ensure fmt succeeds: !"make fmt"
  - If it fails, fix until fmt succeeds. Even if it looks unrelated, you must fix.
- Ensure vet succeeds: !"make vet"
  - If it fails, fix until vet succeeds. Even if it looks unrelated, you must fix.
- Commit if there's any changes
