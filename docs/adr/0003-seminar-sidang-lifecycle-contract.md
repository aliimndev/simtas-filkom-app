# ADR-0003 — Canonical Seminar and Sidang Lifecycle Contract

- Status: Accepted
- Date: 2026-08-28
- Scope: Seminar, Sidang, Revisi, and Tugas Akhir lifecycle integration

## Context

The database already contains `seminars` and `thesis_defenses`, but the application has no Hono route/service implementation for these modules. The existing database constraints, legacy Go implementation, and Svelte placeholder pages use different status names:

- Seminar database accepts `pending`, `scheduled`, `completed`, `passed`, and `failed`.
- Sidang database accepts `pending`, `scheduled`, `completed`, `passed`, `failed`, and `revision_required`.
- Placeholder UI filters use `submitted` and `in_revision`.
- The legacy implementation uses `pending → scheduled → passed/failed` for Seminar and `pending → scheduled → passed/failed/revision_required` for Sidang.

A single canonical contract is required before implementing the Bun + Hono modules.

## Decision

The domain and API use the database vocabulary as the canonical status vocabulary. `submitted` and `in_revision` are not persisted or exposed as API status values; they may only appear as historical/UI wording if needed during a transition.

### Seminar

```text
pending → scheduled → passed
        ↓           → failed
     cancelled
```

- `pending`: Mahasiswa has submitted a Seminar request and it awaits scheduling.
- `scheduled`: the schedule and Penguji assignment are present.
- `passed`: all assigned Penguji have submitted scores and the weighted final score is at least 60.
- `failed`: all assigned Penguji have submitted scores and the weighted final score is below 60.
- `cancelled`: Kaprodi or Admin Fakultas cancelled a `pending` or `scheduled` attempt with a mandatory reason; the attempt is terminal and the parent Thesis remains `seminar_ready`.

`completed` is retained temporarily in the database constraint for backward compatibility with existing data, but new application code must not create or transition to it. A later data audit may remove the unused value through an expand-contract migration.

### Sidang

```text
pending → scheduled → passed
        ↓           → revision_required
     cancelled      → failed
```

- `pending`: Mahasiswa has submitted a Sidang request and it awaits scheduling.
- `scheduled`: the schedule and Penguji assignment are present.
- `passed`: all assigned Penguji have submitted scores and the weighted final score is at least 75.
- `revision_required`: all assigned Penguji have submitted scores and the weighted final score is 60–74; Revisi must be completed before graduation.
- `failed`: all assigned Penguji have submitted scores and the weighted final score is below 60.
- `cancelled`: Kaprodi or Admin Fakultas cancelled a `pending` or `scheduled` attempt with a mandatory reason; the attempt is terminal and the parent Thesis remains `defense_ready`.

`completed` is retained temporarily in the database constraint for the same backward-compatibility reason and is not written by new code. The `cancelled` value requires a schema migration before it can be persisted.

### Revisi and resubmission

- Seminar Revisi is represented by reviewer/examiner notes and revised Documents. Seminar status remains `passed` after the Seminar result is finalized; Revisi is not a Seminar status.
- A failed Seminar or Sidang attempt remains immutable as a historical attempt. A failed Seminar leaves the parent Thesis in `seminar_ready`, allowing a later new request when active-attempt rules permit it; a failed Sidang leaves the parent Thesis in `defense_ready`.
- Surat Tugas lifecycle is `draft → issued` or `draft → cancelled`; `issued → cancelled` is allowed only to Kaprodi/Admin Fakultas with a mandatory cancellation reason, and `cancelled` is terminal. An issued letter used by a `scheduled` Sidang cannot be cancelled directly; the Sidang must first go through a separate authorized administrative cancellation procedure. Issued letters cannot return to draft or be edited.
- Each Thesis may have only one active Sidang attempt (`pending` or `scheduled`) at a time; finalized `passed`, `failed`, and `revision_required` attempts remain historical.
- Each Sidang attempt may have only one active Surat Tugas (`draft` or `issued`) at a time; `cancelled` letters remain historical. PostgreSQL partial unique indexes must enforce both invariants so concurrent requests cannot create duplicates.
- A passed Seminar or Sidang cannot be rescheduled, rescored, or moved backward.
- Finalizing a passed Seminar records the `seminar_done` milestone in audit history and moves the parent Thesis directly from `seminar_ready` to `defense_ready` in the same transaction; `seminar_done` is not a persistent waiting status.
- Finalizing a Sidang with `passed` or `revision_required` moves the parent Thesis to `defense_done`; finalizing a failed attempt does not change the parent Thesis state.

### Policy values

The v1.0 implementation adopts the established legacy policy values:

- Seminar scheduling requires at least 3 days of lead time.
- Sidang scheduling requires at least 7 days of lead time.
- Each event requires at least 2 active Dosen Penguji.
- Seminar pass threshold: 60.
- Sidang thresholds: below 60 = `failed`; 60–74 = `revision_required`; 75 or higher = `passed`.
- Fixed rubric for both events:
  - Presentasi: 30%
  - Penguasaan Materi: 30%
  - Kualitas Naskah: 25%
  - Kemampuan Menjawab: 15%

Weights are fixed in v1.0 and must sum to 100. Dynamic rubric configuration is out of scope.

### API and UI representation

API responses use the canonical values exactly:

```text
pending           → Diajukan
scheduled         → Terjadwal
passed            → Lulus
failed            → Tidak Lulus
revision_required → Perlu Revisi
cancelled         → Dibatalkan
```

The Bahasa Indonesia labels belong in the UI presentation layer. API contracts, database values, audit events, and tests use the canonical English snake-case values above.

## Consequences

- Svelte placeholder filters must be updated to use `pending`, `scheduled`, `passed`, `failed`, and `revision_required` where applicable.
- New Hono routes/services must reject invalid transitions and must not write `completed`, `submitted`, or `in_revision`.
- Existing rows with `completed` require a pre-implementation data audit. If none exist, a migration may tighten the database constraint; if any exist, they remain explicitly legacy until a separate manual resolution. New endpoints must never write, expose, or silently map `completed` to another outcome.
- Sidang cancellation requires a migration adding `cancelled` to the database constraint and a mandatory cancellation-reason field; Seminar cancellation has the same requirement. Both are audited and cannot be applied to finalized attempts.
- Penguji may enter or update only their own scores before finalization; Kaprodi or Admin Fakultas alone may finalize a Seminar/Sidang.
- Seminar/Sidang scoring and finalization need transaction/race protection so the final result is produced once after all assigned Penguji submit.
- Every critical lifecycle mutation must write an audit event in the same business transaction and create in-app notifications for relevant actors. Email is a non-blocking side effect; delivery failures are recorded in `email_logs` or the retry queue.
- Individual score edits are audited but do not notify every stakeholder; the finalized result does notify the relevant actors.
- Backend authorization is authoritative: every related ID and role scope must be verified server-side. Unauthorized access returns `403`; a missing resource returns `404` without leaking cross-user data.
- Mahasiswa owns only their Thesis submissions and documents; Penguji accesses only assigned Sidang records and scores; Dosen Pembimbing reviews only final documents for Thesis they supervise; Kaprodi/Admin Fakultas may operate across records within their authority.
- Finalization must verify every assigned Penguji has completed every fixed rubric component. After finalization, scores, outcome, and examiner assignments are immutable; any administrative correction requires a separate audited procedure.
- Every lifecycle mutation must write an audit event and the required notification side effects.
- Seminar scoring/finalization is implemented before the Sidang module because `defense_ready` must be produced by a verified Seminar outcome, not by a manually set Thesis status.
- The Seminar API contract is intentionally small: `POST /api/v1/theses/:thesisId/seminars` submits a new attempt, `GET /api/v1/seminars` and `GET /api/v1/seminars/:id` provide role-scoped reads, `PUT /api/v1/seminars/:id/schedule` schedules and assigns Penguji, `PUT /api/v1/seminars/:id/scores` atomically upserts all four fixed rubric components for the authenticated assigned Penguji, and `POST /api/v1/seminars/:id/finalize` closes the result for Kaprodi/Admin Fakultas. Retry uses the submission endpoint; no separate retry endpoint is needed.
- Score writes reject missing or duplicate rubric components, unknown component names, changed weights, and scores outside 0–100. Optional per-component Penguji notes are stored with each score, do not affect the weighted result, are editable only before finalization, and become immutable after finalization; the schema migration must add the notes field to Seminar/Sidang score tables.
- This decision does not specify every endpoint or UI layout; those belong in the implementation tickets.

## Consequences

- Graduation is an atomic operation: the system must verify the eligible Sidang outcome and approved latest `final_thesis`, create the single Thesis Archive, and move the Thesis to `graduated` in one transaction.
- The Dosen Pembimbing terkait is the primary reviewer for the latest `final_thesis`; Kaprodi or Admin Fakultas may perform an authorized administrative review.
- Approval of the latest `final_thesis` completes Revisi and opens the archive/graduation gate.
- The archive metadata is part of the graduation record, not an optional post-graduation attachment.
- A Thesis with an existing archive or `graduated` status cannot be archived again or moved backward.

## Out of scope

Dynamic grading weights, multiple concurrent active attempts, SIAKAD integration, and changes to the broader Tugas Akhir state machine beyond the existing Seminar/Sidang integration are outside this ADR.
