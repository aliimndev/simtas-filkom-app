# Simtas Filkom App — Domain Context

## What this is

`simtas-filkom-app` is the frontend web application for the **SIMTAS** (Sistem Manajemen Tugas Akhir / Thesis Management System) used by the **Faculty of Computer Science (Filkom)**. It lets students, advisors, and examiners manage the thesis lifecycle — proposal submission, progress tracking, seminar scheduling, and final defense — through a role-based interface.

## Core concepts

| Concept | Meaning in this repo |
|---------|----------------------|
| SIMTAS | Thesis management system for Filkom |
| Tugas Akhir (TA) | Final-year thesis / final project |
| Mahasiswa | Student — submits and tracks their thesis |
| Dosen Pembimbing | Thesis advisor / supervisor |
| Penguji | Examiner / committee member for seminars and defense |
| Kaprodi | Head of study program — approves flow and resolves escalations |
| Proposal | Initial thesis proposal document submitted for review |
| Seminar | Mid-term and final thesis presentations / defenses |
| Revisi | Revision cycle triggered by advisor/examiner feedback |
| Gate Sidang | Conditions that must be satisfied before a Thesis can enter the defense stage: the Seminar is passed, all assigned Penguji have submitted scores, and the final score reaches at least 60/100. |
| Surat Tugas | Official assignment letter for one Attempt Sidang, identified by its number, issue date, file, issuer, and lifecycle status. `Kaprodi` or `Admin Fakultas` may create and issue it; `draft → issued` is the normal path, while `cancelled` is terminal. Once issued, its identifying data is immutable; replacement requires cancellation with a reason followed by a new letter. An issued Surat Tugas is required before the Sidang can be scheduled. Cancelled letters remain historical, while each retry requires a new letter. |
| Pengajuan Sidang | An explicit request by a Mahasiswa to start one Sidang attempt after the Seminar has passed. The request creates a pending attempt and may be rejected when the final required Sidang document is not approved or another active attempt exists. |
| Dokumen Sidang | The latest `defense_doc` submitted for conducting a Sidang. It must be approved before a Mahasiswa can submit a Sidang request. |
| Dokumen Final | The latest `final_thesis` submitted after Sidang. It must be approved before a Thesis can be archived and marked `graduated`. |
| Penjadwalan Sidang | The administrative action that assigns a `scheduled` time, room, and at least two active Penguji to a pending Sidang. It requires an issued Surat Tugas, seven days of lead time, and no room or Penguji conflict. |
| Rubrik Sidang | The fixed four-component assessment used by every assigned Penguji: Presentasi 30%, Penguasaan Materi 30%, Kualitas Naskah 25%, and Kemampuan Menjawab 15%. Each component is scored from 0 to 100, and the finalized result is the weighted average across all Penguji. |
| Finalisasi Sidang | The one-time action that closes scoring after every assigned Penguji has completed every rubric component and determines the Sidang outcome. Only Kaprodi or Admin Fakultas may finalize it; Penguji may only enter scores before finalization. A finalized score, outcome, and examiner assignment cannot be changed. A failed result leaves the Thesis in `defense_ready`; `passed` or `revision_required` moves it to `defense_done`. |
| Arsip Tugas Akhir | The single permanent record of a completed Thesis, containing the approved final thesis file and publication metadata such as bilingual abstracts, keywords, and graduation year. |
| Kelulusan | The irreversible Thesis outcome produced atomically when an eligible final Sidang outcome has its latest `final_thesis` approved and its Arsip Tugas Akhir created. |
| Review Dokumen Final | The post-Sidang approval of the latest `final_thesis`. Dosen Pembimbing terkait is the primary reviewer; Kaprodi or Admin Fakultas may perform an authorized administrative review. Approval completes Revisi and opens the archive/graduation gate. |
| Attempt Sidang | One immutable Sidang execution record for a Thesis. A failed or cancelled attempt remains history; a new attempt may be submitted without a v1.0 count limit, but only one `pending` or `scheduled` attempt may be active and each new attempt requires a new Surat Tugas and approved latest Dokumen Sidang. |
| Pembatalan Sidang | An administrative cancellation of a `pending` or `scheduled` Sidang by Kaprodi or Admin Fakultas, with a mandatory reason and audit record. It is terminal, leaves the Thesis `defense_ready`, and requires a new attempt and Surat Tugas for any retry. |
| Audit Lifecycle | The mandatory audit trail for every critical lifecycle mutation. Audit insertion is part of the business transaction; in-app notifications target relevant actors, while email delivery is non-blocking and failures are recorded for retry. |
| Backend Authorization | The server-side enforcement of Thesis ownership and role scope: Mahasiswa, Penguji, and Dosen Pembimbing may access only their related records, while Kaprodi/Admin Fakultas may act across records within their authority. |
| Attempt Seminar | One immutable Seminar execution record for a Thesis. A failed or cancelled attempt remains history while the Thesis stays `seminar_ready`; a new request requires an approved latest `seminar_doc`, and only one `pending` or `scheduled` attempt may be active. A passed Seminar cannot be resubmitted. |
| Pembatalan Seminar | An administrative cancellation of a `pending` or `scheduled` Seminar by Kaprodi or Admin Fakultas, with a mandatory reason and audit record. It is terminal, leaves the Thesis `seminar_ready`, and requires a new attempt for any retry. |
| Finalisasi Seminar | The one-time action that closes Seminar scoring. A passed result records the `seminar_done` milestone in audit history and moves the Thesis directly to `defense_ready`; `seminar_done` is not a persistent waiting status. |
| Catatan Penguji | Optional feedback attached to each rubric component by its assigned Penguji. It does not affect the score, is editable only before finalization, and becomes immutable evidence for Revisi after finalization. |

- **Mahasiswa** — submit proposal, upload drafts, schedule seminars, respond to revisions.
- **Dosen Pembimbing** — review drafts, approve/reject proposals, schedule seminars, request revisions.
- **Penguji** — evaluate seminar/defense presentations, submit scores and remarks.
- **Kaprodi** — oversee the entire flow, approve assignment letters, resolve exceptions.
- **Admin** — system administration, user management, reporting.

## Key flows

1. **Proposal flow**: mahasiswa submits proposal → Kaprodi approves → Dosen Pembimbing assigned.
2. **Seminar flow**: proposal approved → mahasiswa schedules seminar → Dosen Pembimbing + Penguji evaluate.
3. **Revisi flow**: feedback given → mahasiswa uploads revised draft → advisor re-reviews.
4. **Sidang / Defense flow**: seminar passed → Surat Tugas issued → final defense → graduation clearance.

## Notes

- This `CONTEXT.md` captures the shared vocabulary; architectural decisions belong in `docs/adr/`. The current release plan lives in `docs/ROADMAP.md`.
- Use these exact terms in issue titles, PR descriptions, and code comments rather than loose synonyms (e.g. use **Proposal** instead of "application", **Dosen Pembimbing** instead of "supervisor", etc.).
