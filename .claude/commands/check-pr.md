---
description: "Check PR review and CI status"
---

- If you cannot identify which PR to check, ask user
- Check PR reviews. make sure to check both comments and inline comments
  - Have golang-pro agent fix any suggestions if it's reasonable
  - If you skip resolving issues or suggestions, you must add reasoning behind that dicision.
  - All issues should be considered. Do not skip issues or suggestions just because it's minor or low priority.
- Check CI status. 
  - Make sure to check what is actually failing in the CI
    - CI has multiple steps in a single job. Identify which step is failing
    - Report the cause before fixing it or running test/lint/format locally
    - DO NOT manually trigger the workflow or create a empty commit to trigger workflow run.
  - Have golang-pro agent fix issues
