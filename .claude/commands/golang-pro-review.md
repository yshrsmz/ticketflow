---
description: "review by golang-pro"
argument-hint: changes | pr
---

- If argument is "changes", have golang-pro subsgent review local diff from parent branch.
- If argument is pr, have golang-pro review changes in PR
  - if you cannot identify which PR, ask user
- After the review, fix all issues and suggestions. Regardless of the priority, fix all issues & suggestions.
  - If it's too big to fix in the current ticket, you can create a new issue only after you got permission from user.
