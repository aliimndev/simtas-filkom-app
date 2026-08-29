# AGENTS.md


## Agent skills

### Issue tracker

GitHub Issues via the `gh` CLI. See `docs/agents/issue-tracker.md`.

### Triage labels

Five canonical triage roles. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context layout. See `docs/agents/domain.md`.

### Agent execution

Follow the sequential self-correction loop in `docs/AGENT-EXECUTION-LOOP.md`; use GitHub Issues as the persistent source of truth.

## Ponytail coding mode

Apply the Ponytail lazy-senior-dev rules from `/home/alee/.config/opencode/.ponytail-active` for every coding task: understand and trace the real flow first, prefer reuse and standard-library/platform solutions, avoid unnecessary abstractions and dependencies, keep the diff minimal, fix bugs at the root cause, and leave one runnable check for non-trivial logic.
