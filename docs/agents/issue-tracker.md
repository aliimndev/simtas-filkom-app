# Issue tracker: GitHub

Issues and PRDs for this repo live as GitHub Issues. Use the `gh` CLI for issue-tracker operations.

## Conventions

- Create an issue: `gh issue create --title "..." --body "..."`.
- Read an issue: `gh issue view <number> --comments` and inspect labels/comments.
- List issues: `gh issue list` with appropriate `--state` and `--label` filters.
- Comment on an issue: `gh issue comment <number> --body "..."`.
- Apply or remove labels: `gh issue edit <number> --add-label "..."` / `--remove-label "..."`.
- Close an issue: `gh issue close <number> --comment "..."`.

The repository is inferred from the current clone's `origin` remote.

## Pull requests as a triage surface

**PRs as a request surface: no.** External pull requests are not included in ordinary triage discovery. An explicitly named PR may still be reviewed directly.

## When a skill says “publish to the issue tracker”

Create a GitHub issue.

## Wayfinding operations

- Map: create one issue labelled `wayfinder:map`.
- Child tickets: create issues linked to the map as GitHub sub-issues when supported; otherwise put `Part of #<map>` at the top of the body and maintain the map task list.
- Blocking: use GitHub native issue dependencies when available. Otherwise record `Blocked by: #<n>, #<n>` at the top of the child issue body.
- Claim: assign the issue to the current user before implementation.
- Resolve: comment with the answer, close the issue, then add a decision pointer to the map.
