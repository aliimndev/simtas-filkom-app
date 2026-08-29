# ADR-0004 — Surat Tugas Required Before Sidang Scheduling

- Status: Accepted
- Scope: Sidang administration and scheduling

A Sidang must have an issued **Surat Tugas** before it can be scheduled. `Surat Tugas` belongs directly to one Attempt Sidang through `surat_tugas.defense_id → thesis_defenses.id`. `Kaprodi` or `Admin Fakultas` may create a draft and issue it. Each attempt may have one active letter; cancelled letters remain historical, and every retry requires a new letter. Surat Tugas is modeled as an auditable domain entity containing at least the letter number, issue date, file, issuer, and lifecycle status `draft`, `issued`, or `cancelled`. Once issued, its identifying data is immutable; a correction or replacement requires cancelling the letter and creating a new one. This keeps the institutional administrative requirement verifiable instead of representing it as an untraceable flag on the Thesis or Sidang.
