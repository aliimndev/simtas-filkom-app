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
                    → failed
```

- `pending`: Mahasiswa has submitted a Seminar request and it awaits scheduling.
- `scheduled`: the schedule and Penguji assignment are present.
- `passed`: all assigned Penguji have submitted scores and the weighted final score is at least 60.
- `failed`: all assigned Penguji have submitted scores and the weighted final score is below 60.

`completed` is retained temporarily in the database constraint for backward compatibility with existing data, but new application code must not create or transition to it. A later data audit may remove the unused value through an expand-contract migration.

### Sidang

```text
pending → scheduled → passed
                    → revision_required
                    → failed
```

- `pending`: Mahasiswa has submitted a Sidang request and it awaits scheduling.
- `scheduled`: the schedule and Penguji assignment are present.
- `passed`: all assigned Penguji have submitted scores and the weighted final score is at least 75.
- `revision_required`: all assigned Penguji have submitted scores and the weighted final score is 60–74; Revisi must be completed before graduation.
- `failed`: all assigned Penguji have submitted scores and the weighted final score is below 60.

`completed` is retained temporarily in the database constraint for the same backward-compatibility reason and is not written by new code.

### Revisi and resubmission

- Seminar Revisi is represented by reviewer/examiner notes and revised Documents. Seminar status remains `passed` after the Seminar result is finalized; Revisi is not a Seminar status.
- Sidang Revisi is represented by `revision_required`, revision notes, revised Documents, and the final thesis approval gate. Graduation may proceed only after the final thesis Document is approved.
- A failed Seminar or Sidang attempt remains immutable as a historical attempt. A new request may be submitted only when the parent Tugas Akhir state and active-attempt rules permit it; the implementation must never overwrite the failed attempt.
- A passed Seminar or Sidang cannot be rescheduled, rescored, or moved backward.
- A `revision_required` Sidang cannot be rescored or rescheduled; its completion is proven by the final thesis approval and graduation flow.

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
```

The Bahasa Indonesia labels belong in the UI presentation layer. API contracts, database values, audit events, and tests use the canonical English snake-case values above.

## Consequences

- Svelte placeholder filters must be updated to use `pending`, `scheduled`, `passed`, `failed`, and `revision_required` where applicable.
- New Hono routes/services must reject invalid transitions and must not write `completed`, `submitted`, or `in_revision`.
- Existing rows with `completed` require compatibility handling; they must not be silently rewritten without a data migration decision.
- Seminar/Sidang scoring and finalization need transaction/race protection so the final result is produced once after all assigned Penguji submit.
- Every lifecycle mutation must write an audit event and the required notification side effects.
- This decision does not specify every endpoint or UI layout; those belong in the implementation tickets.

## Out of scope

Dynamic grading weights, multiple concurrent active attempts, SIAKAD integration, and changes to the broader Tugas Akhir state machine beyond the existing Seminar/Sidang integration are outside this ADR.
