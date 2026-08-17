# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

SIMTAS is a multi-role web platform; all confirmed roles carry equal weight as primary users (per maintainer confirmation), each acting on the shared thesis workflow from their own dashboard:

- **Mahasiswa (Student)** — proposes thesis titles, uploads documents, tracks guidance/consultation progress, and follows the pipeline to seminar, defense (sidang), and graduation archive.
- **Dosen Pembimbing (Supervisor)** — validates guidance logs, reviews uploaded documents, and accompanies students through the pipeline.
- **Dosen Penguji (Examiner)** — reviews documents, is assigned to seminars/defenses, and inputs scores using the fixed-weight rubric.
- **Kaprodi (Head of Program)** — reviews and approves proposed titles, assigns supervisors/examiners, monitors progress, and manages scheduling.
- **Admin Fakultas (Faculty Admin)** — manages users (CRUD, CSV/Excel import, activation, password reset), system configuration, academic-year management, and master data.

## Product Purpose

SIMTAS (Sistem Manajemen Tugas Akhir Skripsi) manages the entire undergraduate thesis (tugas akhir/skripsi) lifecycle for the Fakultas Ilmu Komputer Universitas Djuanda — from title submission, review and supervisor assignment, guidance and document upload/approval, seminar and defense scheduling and scoring, revision, through graduation status and final digital archive.

It replaces a manual, paper-based process that suffered from lost physical documents, delayed information, hard-to-track progress, inconsistent procedures/scoring, high administrative load, and weak reporting/audit. Success is measured by full adoption, reduced approval time, complete thesis records, high user satisfaction, and high system availability.

## Positioning

A single centralized, digital, end-to-end platform that binds all five stakeholder roles into one accountable workflow with a full audit trail — something each role previously managed piecemeal on paper. Its meaningfully different mechanism is the complete, state-tracked, audited thesis pipeline in one system, with consistent (fixed-weight) scoring.

## Operating Context

- Indonesia-language product (Bahasa Indonesia is the sole interface language in v1.0).
- Runs in a campus environment; data is per-student and must remain under faculty control.
- Deployment targeted for on-campus production (see `infrastructure/`, `deploy/`, and `docs/PRODUCTION-READINESS-REVIEW.md`).
- Backend Go (Gin, GORM, Clean Architecture) + PostgreSQL 16, with Supabase Storage for PDF/archive objects, Resend for email notifications, and JWT auth. Frontend Next.js + React + TypeScript + Tailwind CSS v4.
- API is versioned at `/api/v1/...` and documented with Swagger/OpenAPI.
- Seven core email-trigger events notify relevant stakeholders; audit log records all critical actions.

## Capabilities and Constraints

Confirmed capabilities (v1.0):

- Digital title/thesis submission with review and approval by Kaprodi.
- Supervisor assignment; guidance/consultation log management; document upload with approval workflow.
- Proposal seminar approval, scheduling, fixed-weight scoring, and post-seminar revision.
- Defense (sidang) submission, scheduling, examiner scoring with fixed weights, final revision, and graduation/yudisium status determination.
- Digital archive of the final thesis.
- Role-based access control across 5 roles; progress dashboards; email notifications (7 events); audit logs; search/filter/pagination; academic-year management; user management with CSV/Excel import and activation.
- Responsive web (no native mobile app in v1.0).

Confirmed constraints and technical facts:

- Scoring uses fixed-weight rubrics (dynamic weight configuration is deferred to a later version).
- Document versioning, SIAKAD integration, WhatsApp notifications, self-registration, video-conference integration, plagiarism checking, e-signature, external API, real-time chat, and advanced analytics/ML are explicitly out of scope for v1.0.
- Storage is S3-API-compatible object storage (Supabase Storage now; swappable via the backend `StorageService` interface without rewriting upload logic).

Undecided (recorded, not invented): exact production storage vendor confirmation, final deployment timeline/versioning specifics beyond v1.0.

## Brand Commitments

- Official name: **SIMTAS FILKOM** (Sistem Manajemen Tugas Akhir Skripsi), Fakultas Ilmu Komputer Universitas Djuanda.
- Institutional identity: FILKOM/Unida blue is used as the accent across the interface.
- Product voice is formal, academic, and administrative (formal Indonesian).
- Version labeling: v1.0 (Development status).

## Evidence on Hand

- `PRD.md` — full product requirements document (workflows, scope, metrics, roles).
- `README.md` — stack, architecture, and setup documentation.
- `docs/` — production readiness review, release roadmap, and operational runbook.
- `frontend/` — implemented Next.js interface (see `frontend/DESIGN.md` for the incumbent visual system) with working routes for all roles.
- `backend/` — Go API implementing the documented workflows.

State of proof: no customer testimonials, benchmarks, or press exist; future work must not fabricate them.

## Product Principles

1. End-to-end accountability: every critical action is tracked and auditable from submission to archive.
2. Consistency over flexibility: standardized procedures and fixed-weight scoring keep evaluations uniform across the faculty.
3. Digital-first: replace manual/paper steps with structured digital workflows that scale.
4. Role-appropriate clarity: each of the five roles sees exactly the state, actions, and notifications relevant to them.
5. Data control and durability: per-student records stay under faculty control, are complete, and are safely archived.

## Accessibility & Inclusion

No product-specific accessibility requirement has been established beyond general web best practices (semantic HTML, keyboard-operable controls, sufficient contrast, responsive layout). The interface is in Bahasa Indonesia; content accessibility should follow standard WCAG-level practice where feasible.

