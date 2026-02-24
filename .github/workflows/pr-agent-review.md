---
on:
  workflow_dispatch:
  pull_request:
    types: [opened, synchronize, reopened]

permissions:
  contents: read
  pull-requests: read
  issues: read

tools:
  github:
    toolsets: [default]

network: defaults

safe-outputs:
  add-comment:
    max: 1
---

# pr-agent-review

The purpose of this workflow is to help maintainers triage pull requests in a concise and consistent way.

## Instructions

1. Read the PR title, description, changed files, and repository context.
2. Focus on change risk for this Go project:
   - API compatibility of the `paths` and `types` packages
   - cross-platform path behavior changes
   - type/struct validity that may impact downstream consumers
3. Verify minimum quality from a reviewer perspective:
   - whether the change needs new tests
   - whether `README.md` documentation should be updated
   - whether there is potential breaking change risk
4. Write a very concise review comment with this format:
   - Summary
   - Risks
   - Recommended actions (maximum 3 points)
5. If there are no significant risks, state that the PR appears safe with a short note.

## Constraints

- Do not make code changes.
- Do not create new issues or pull requests.
- Only post one actionable comment.
