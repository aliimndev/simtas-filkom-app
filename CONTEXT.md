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
| Surat Tugas | Official assignment letter required before defense |

## User roles

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

- The repo name and existing `docs/` phase folders (`phase-1-foundation`, `phase-2-core-backend`, etc.) indicate this is a multi-phase project. This `CONTEXT.md` captures the shared vocabulary; per-phase decisions belong in `docs/adr/`.
- Use these exact terms in issue titles, PR descriptions, and code comments rather than loose synonyms (e.g. use **Proposal** instead of "application", **Dosen Pembimbing** instead of "supervisor", etc.).
