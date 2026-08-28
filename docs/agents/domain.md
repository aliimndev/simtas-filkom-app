# Domain documentation

This repository uses a single-context domain layout.

## Before exploring

Read the root `CONTEXT.md` when it exists. Also read relevant records in `docs/adr/` before changing behavior in an area governed by an architectural decision. If these files are absent, proceed without creating them merely for setup.

## Layout

```text
/
├── CONTEXT.md
├── docs/adr/
└── apps/ and packages/
```

Domain terminology in issues, plans, code, and tests should follow the vocabulary established in `CONTEXT.md`. If a needed concept is not defined there, note the terminology gap and use the project's existing code vocabulary until it is resolved.

If a change conflicts with an ADR, call out the conflict explicitly rather than silently overriding the decision.
