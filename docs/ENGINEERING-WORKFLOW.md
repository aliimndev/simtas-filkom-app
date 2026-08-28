# SIMTAS FILKOM — Engineering Workflow A–Z

Dokumen ini adalah workflow standar untuk membangun, mengubah, merilis, dan mengoperasikan SIMTAS FILKOM. Workflow ini memakai skill engineering Matt Pocock yang tersedia di environment lokal dan disesuaikan dengan domain SIMTAS, bukan workflow generik.

## 0. Tujuan dan batasan

### Tujuan

Setiap perubahan harus:

1. menyelesaikan kebutuhan nyata salah satu role SIMTAS;
2. menjaga state machine Tugas Akhir tetap valid;
3. menjaga auditability, keamanan akses, dan konsistensi data;
4. dapat diverifikasi melalui test pada seam publik;
5. dapat dirilis atau di-rollback dengan aman.

### Stack yang menjadi baseline

- Backend: Bun + Hono + TypeScript.
- Frontend: SvelteKit + Svelte + TypeScript.
- Data: PostgreSQL + Drizzle schema/migrations.
- Storage: object storage S3-compatible melalui boundary storage service.
- Email: provider email melalui boundary email service.
- Local/CI: Bun, PostgreSQL, Docker Compose, GitHub Actions.
- Frontend deployment: Vercel.
- API production: Docker/VPS atau service Bun yang setara.

### Vocabulary wajib

Gunakan istilah domain di `CONTEXT.md` dan `PRODUCT.md`: `Mahasiswa`, `Dosen Pembimbing`, `Dosen Penguji`, `Kaprodi`, `Admin Fakultas`, `Proposal`, `Seminar`, `Sidang`, `Revisi`, `Surat Tugas`, `Tugas Akhir`, dan `Arsip`.

Jangan mengganti istilah tersebut dengan istilah teknis yang kabur di issue, API contract, test name, atau UI copy. Contoh: gunakan `Dosen Pembimbing`, bukan hanya `supervisor`; gunakan `Proposal`, bukan `application`.

---

## 1. Peta skill yang digunakan

| Kebutuhan | Skill utama | Output |
| --- | --- | --- |
| Menentukan flow yang tepat | `ask-matt` | pilihan flow yang sesuai |
| Pekerjaan sangat besar dan masih berkabut | `wayfinder` | map issue + decision tickets |
| Memperjelas istilah dan aturan domain | `domain-modeling`, `ubiquitous-language` | glossary/context + ADR bila perlu |
| Menguji ide UI/state sebelum dibangun | `prototype`, `frontend-design` | prototype yang dapat direaksi |
| Menyusun spec dari percakapan | `to-spec` | GitHub issue berlabel `ready-for-agent` |
| Memecah spec menjadi vertical slices | `to-tickets` | issue tracer-bullet + dependency edges |
| Menulis fitur dari ticket | `implement` | perubahan code + tests |
| Perubahan test-first | `tdd` | red-green-refactor di seam yang disepakati |
| Bug sulit atau regresi performa | `diagnosing-bugs`, `systematic-debugging` | diagnosis berbasis bukti + fix |
| Review branch/PR | `code-review`, `review` | review Standards dan Spec |
| QA percakapan dan issue baru | `qa`, `triage` | issue terklasifikasi dan agent-ready |
| Refactor besar | `request-refactor-plan`, `codebase-design` | plan bertahap dengan seam yang lebih dalam |
| Validasi final | `verification-before-completion` | evidence check sebelum claim/merge |
| Handoff lintas session | `handoff`, `claude-handoff` | context ringkas yang dapat dilanjutkan |

### Aturan pemilihan flow

- Jangan memakai `wayfinder` untuk fitur kecil yang bisa selesai dalam satu session.
- Jangan memakai `triage` untuk ticket yang baru dibuat oleh `to-tickets`; ticket tersebut sudah `ready-for-agent`.
- Jangan langsung `implement` jika istilah domain, scope, atau acceptance criteria masih ambigu.
- Jangan menulis test pada seam yang belum disepakati; konfirmasi seam sebelum siklus TDD.
- Jangan claim selesai tanpa fresh verification evidence.

---

## 2. Lifecycle A–Z

### A — Align: mulai dari outcome

Tulis outcome dalam bahasa pengguna, bukan solusi teknis.

Format minimum:

```text
Untuk [role], ketika [kondisi], sistem harus memungkinkan [perilaku]
agar [hasil bisnis/domain].
```

Contoh:

```text
Untuk Mahasiswa, setelah Proposal disetujui Kaprodi, sistem harus memungkinkan
Mahasiswa mengajukan jadwal Seminar agar proses review tidak lagi bergantung pada
pesan pribadi atau dokumen kertas.
```

Tentukan juga apakah pekerjaan berupa:

- bug/regresi;
- fitur domain baru;
- perubahan policy/state machine;
- refactor;
- perubahan operasional/infrastruktur;
- hardening keamanan.

### B — Boundaries: tetapkan scope

Tentukan:

- role yang terdampak;
- state Tugas Akhir yang terdampak;
- modul data/API/UI yang terdampak;
- in-scope dan out-of-scope;
- kebutuhan migrasi dan rollback;
- apakah perubahan menyentuh data production.

V1 tetap mengikuti batasan `PRODUCT.md`: tidak menambahkan SIAKAD integration, WhatsApp notification, self-registration, video conference, plagiarism checker, e-signature, chat realtime, atau advanced ML tanpa keputusan scope baru.

### C — Context: baca konteks sebelum code

Sebelum eksplorasi:

1. baca `CONTEXT.md`;
2. baca `PRODUCT.md`;
3. baca `docs/ROADMAP.md`;
4. baca ADR yang relevan di `docs/adr/`;
5. baca `docs/agents/issue-tracker.md`, `triage-labels.md`, dan `domain.md`;
6. cek branch, status, remote, dan CI terbaru.

Skill: `domain-modeling` bila ada istilah atau aturan yang belum tajam.

### D — Discover: cari implementasi yang sudah ada

Lakukan pencarian berdasarkan konsep domain, bukan hanya kata-kata request:

- route/API yang sudah ada;
- service/use case;
- schema dan migration;
- state/status transition;
- UI route dan component;
- notification/audit behavior;
- test parity atau integration test.

Jika perilaku sudah ada, jangan membuat duplikasi. Jika kode dan domain statement berbeda, hentikan dan catat contradiction sebelum memilih salah satu.

### E — Explore architecture and seams

Pilih seam publik tertinggi yang bisa mengamati behavior:

- API request/response untuk flow backend;
- halaman/dashboard atau component interaction untuk flow UI;
- CLI/script boundary untuk operasional;
- database migration hanya diverifikasi melalui behavior API bila memungkinkan.

Utamakan satu seam yang mencakup sebanyak mungkin behavior. Hindari test terhadap private function, query internal, atau mock collaborator internal.

Skill: `codebase-design` jika boundary terlalu dangkal atau module sulit diuji.

### F — Fog decision

Jika pekerjaan lebih besar dari satu session atau keputusan utama belum jelas, gunakan `wayfinder`.

Wayfinder membuat:

- satu map issue berlabel `wayfinder:map`;
- decision tickets bertipe `research`, `prototype`, `grilling`, atau `task`;
- dependency edges;
- frontier ticket yang dapat diambil berikutnya.

Wayfinder merencanakan keputusan, bukan langsung mengimplementasikan seluruh destination. Satu session tidak boleh menyelesaikan lebih dari satu decision ticket, kecuali research tickets.

### G — Grill the unknowns

Untuk ketidakpastian yang berpengaruh pada scope atau model:

- gunakan `grilling` untuk bertanya satu per satu;
- gunakan `batch-grill-me` bila semua frontier question perlu ditampilkan sekaligus;
- gunakan `domain-modeling` untuk edge case dan istilah;
- gunakan `grill-with-docs` bila keputusan harus langsung direkam ke ADR/glossary.

Pertanyaan wajib untuk flow baru:

1. siapa actor yang boleh memulai, menyetujui, menolak, atau membatalkan?
2. state sebelum dan sesudah action apa?
3. apakah action idempotent?
4. apa yang terjadi jika actor tidak aktif atau assignment berubah?
5. data apa yang wajib diaudit?
6. notifikasi apa yang wajib dikirim dan kapan?
7. apakah rejection membuka Revisi atau menghentikan flow?
8. apa perilaku terhadap deadline, zona waktu, dan academic year?

### H — Harden the model

Sebelum implementasi, stress-test dengan scenario konkret:

- dua actor melakukan approval bersamaan;
- user mencoba mengakses record milik Mahasiswa lain;
- token/session sudah dicabut;
- file upload gagal setelah metadata tersimpan;
- reviewer berubah setelah jadwal dibuat;
- Proposal ditolak lalu diajukan ulang;
- satu Mahasiswa memiliki lebih dari satu Tugas Akhir aktif;
- academic year berganti saat flow berjalan;
- email provider gagal;
- migration dijalankan dua kali;
- rollback deploy setelah migration.

Skill: `harden` untuk mencari loophole policy/strategy, `diagnosing-bugs` untuk behavior yang sudah gagal.

### I — Investigate external facts

Jika keputusan membutuhkan fakta di luar repo, buat `research` ticket dan gunakan `research` skill. Sumber utama harus diprioritaskan. Catat:

- pertanyaan;
- sumber dan tanggal;
- fakta yang terverifikasi;
- keputusan yang dipengaruhi;
- hal yang belum terverifikasi.

Jangan mengirim source code, secret, atau data user ke endpoint eksternal.

### J — Justify irreversible decisions

Buat ADR hanya bila keputusan:

1. sulit dibalik;
2. akan membingungkan future maintainer tanpa konteks;
3. merupakan hasil trade-off nyata.

Contoh yang layak ADR:

- perubahan state machine Tugas Akhir;
- strategi data migration expand-contract;
- pemisahan deployment API Bun dan frontend Vercel;
- policy single active academic year;
- retention/archival data dan audit log.

### K — Keep spec user-facing

Gunakan `to-spec` setelah percakapan cukup jelas. Spec minimal berisi:

- Problem Statement;
- Solution;
- user stories panjang untuk setiap role;
- implementation decisions tanpa bergantung pada path file;
- testing decisions dan seam;
- out of scope;
- further notes.

Spec harus menjawab behavior. Detail implementation yang mudah berubah jangan dikunci tanpa alasan.

### L — List vertical slices

Gunakan `to-tickets` untuk memecah spec menjadi tracer bullets.

Setiap ticket harus:

- lengkap dari schema/data hingga API dan UI bila memang terdampak;
- demoable/verifiable sendiri;
- muat dalam satu context window;
- menyatakan acceptance criteria;
- menyatakan `Blocked by` hanya untuk dependency nyata;
- berlabel `ready-for-agent`.

Jangan membuat ticket horizontal seperti "buat semua API" lalu "buat semua UI".

### M — Map module dependencies

Urutan default modul SIMTAS saat ini:

```text
users
  ↓
roles
  ↓
academic_years
  ↓
theses / Proposal
  ↓
title_change_requests
  ↓
consultation_logs
  ↓
seminars
  ↓
thesis_defenses
  ↓
documents
  ↓
thesis_archives
  ↓
notifications → email_logs → audit_logs → dashboard
```

Urutan ini bukan hukum mutlak. Ticket boleh paralel bila tidak berbagi invariant, migration conflict, atau contract dependency. `notifications`, `email_logs`, `audit_logs`, dan `dashboard` sering menjadi cross-cutting dependency dan harus diberi blocking edge yang nyata.

### N — Name the contract

Sebelum code, kunci contract:

- HTTP method dan path versioned `/api/v1/...`;
- actor/role authorization;
- request validation;
- response shape dan error envelope;
- status code;
- pagination/filter/sort;
- audit action;
- notification event;
- migration compatibility;
- UI loading/empty/error/success states.

Contract harus memakai istilah domain dan tidak expose secret atau data lintas role.

### O — Outline tests before implementation

Catat seam dan behavior test sebelum menulis test:

- API integration test untuk happy path dan authorization;
- unit test hanya untuk pure logic/public utility;
- UI test untuk user-visible interaction;
- migration smoke test untuk schema/data compatibility;
- smoke test health dan critical route setelah deploy.

Kriteria test baik:

- menguji behavior eksternal;
- independen dari implementasi internal;
- nama test berupa requirement;
- expected value berasal dari source of truth yang independen;
- tidak hanya menguji status 200 tanpa payload/side effect penting.

### P — Prototype uncertain behavior

Untuk UI atau state model yang belum jelas, gunakan `prototype` sebelum production implementation.

Prototype boleh throwaway dan tidak perlu production polish. Tujuannya menjawab satu pertanyaan, misalnya:

- bagaimana Mahasiswa melihat timeline Proposal → Seminar → Sidang;
- bagaimana Revisi ditampilkan setelah review;
- bagaimana Kaprodi menangani konflik jadwal;
- bagaimana empty/loading/error state dibaca dalam Bahasa Indonesia.

Untuk visual baru, gunakan `frontend-design`; hindari template UI generik.

### Q — Red: TDD first

Mulai dari satu behavior paling kecil:

1. tulis satu test pada seam yang disepakati;
2. jalankan dan pastikan **red** karena alasan yang benar;
3. implement minimal;
4. jalankan test menjadi **green**;
5. ulangi untuk behavior berikutnya;
6. refactor setelah slice selesai dan tetap green.

Jangan menulis seluruh test suite di awal. Jangan menguji private implementation.

### R — Realize one vertical slice

Implement satu slice end-to-end:

```text
migration/schema (bila perlu)
  → service/domain rule
  → Hono route/contract
  → Svelte UI state/action
  → audit/notification
  → integration/UI test
```

Pertahankan slice kecil. Jalankan typecheck dan test file terkait secara berkala, bukan hanya di akhir.

Skill: `implement`, dengan `tdd` pada seam yang disepakati.

### S — Secure by default

Setiap endpoint baru wajib mengecek:

- authentication;
- role authorization;
- ownership/scoping terhadap Tugas Akhir;
- input validation dan size limit;
- rate limit bila sensitif;
- CORS policy;
- token/session revocation;
- audit action untuk critical mutation;
- error tidak membocorkan detail internal;
- file type/size/storage path bila upload.

Untuk setiap role, uji minimal: allowed actor, forbidden actor, unauthenticated actor, unknown resource, inactive/revoked identity.

### T — Test the transition

Untuk perubahan state machine, test semua transition yang diizinkan dan ditolak.

Minimal matrix:

| Area | Harus diverifikasi |
| --- | --- |
| Proposal | submit, approve, reject, re-submit bila diizinkan |
| Dosen Pembimbing | assign, review, approve/reject dokumen, validasi konsultasi |
| Seminar | request, approve, schedule, evaluate, revision |
| Sidang | eligibility, Surat Tugas, schedule, scoring, final revision |
| Archive | graduation/clearance, final document, immutable/archive access |
| Academic year | exactly-one-active policy dan behavior saat berganti |

### U — Understand failure behavior

Untuk setiap external dependency, tentukan failure policy:

- database unavailable → health `503`, request gagal aman;
- email gagal → critical mutation tidak boleh diam-diam hilang; gunakan retry/log sesuai contract;
- storage gagal → jangan menyisakan metadata yang terlihat sebagai upload sukses;
- duplicate request → idempotent atau `409` yang terdokumentasi;
- concurrent approval → transaction/constraint mencegah state invalid;
- stale token → `401` dan tidak ada side effect.

### V — Verify locally

Perintah standar dari root:

```bash
bun install --frozen-lockfile
bun run typecheck
bun run check
bun run test
bun run build
```

Untuk API integration dengan PostgreSQL:

```bash
docker compose up -d postgres
DATABASE_URL=postgres://postgres:postgres@localhost:5433/simtas \
  bun run --cwd apps/api test
```

Untuk perubahan kecil, jalankan test file target lebih dahulu lalu suite penuh. Catat command, exit code, dan jumlah test dalam PR.

### W — Write the review evidence

Sebelum PR/merge:

1. jalankan `git diff --check`;
2. baca diff dan list commit;
3. jalankan `code-review` terhadap fixed point (`main` atau merge-base);
4. review dua axis terpisah: Standards dan Spec;
5. perbaiki blocker dan rerun verification;
6. gunakan `verification-before-completion` sebelum claim selesai.

Review tidak boleh hanya bergantung pada test pass; cek juga scope creep, authorization, migration safety, dan domain language.

### X — eXecute CI gates

CI wajib memverifikasi:

- API + web typecheck;
- API/web/shared unit tests;
- API integration tests dengan PostgreSQL;
- web build;
- dependency/security audit.

Integration test yang memakai global database invariant harus serial atau memiliki database/schema terisolasi. Jangan menerima test yang hanya lolos karena urutan kebetulan.

### Y — Yield a reviewable PR

Aturan branch/PR:

- branch feature dari `develop` atau branch kerja yang ditentukan tim;
- satu logical change per PR;
- commit convention: `feat:`, `fix:`, `test:`, `docs:`, `chore:`;
- PR body menyebut problem, behavior, scope, migration, rollback, tests, dan screenshots untuk UI;
- jangan commit secret, `.env.local`, generated output, atau data pribadi;
- jangan push langsung ke `main` kecuali maintainer memang meminta;
- squash/merge mengikuti policy repository, tanpa amend commit orang lain.

### Z — Zero-downtime release and rollback

Release flow normal:

```text
issue/spec approved
  → ready-for-agent ticket
  → feature branch
  → TDD vertical slice
  → local verification
  → PR
  → CI green + code review
  → merge to main
  → Vercel Production deployment
  → smoke test
  → monitor
```

Untuk perubahan API/database yang berisiko, gunakan expand-contract:

1. expand: tambah schema/contract yang backward-compatible;
2. migrate callers: pindahkan API/UI/worker secara bertahap;
3. observe: monitor error, latency, audit, dan data;
4. contract: hapus bentuk lama hanya setelah semua caller hilang.

Rollback harus menjawab:

- commit/image yang dikembalikan;
- apakah migration backward-compatible;
- bagaimana mengembalikan API dan web secara konsisten;
- apakah perlu restore backup;
- siapa yang menyetujui rollback;
- bagaimana memverifikasi health setelah rollback.

---

## 3. Workflow per jenis pekerjaan

### Fitur domain baru

```text
ask-matt
→ domain-modeling
→ (wayfinder bila besar)
→ to-spec
→ to-tickets
→ tdd + implement
→ code-review
→ CI
→ deploy + smoke test
```

### Bug/regresi

```text
diagnosing-bugs
→ reproduce with a failing regression test
→ minimal fix
→ targeted verification
→ code-review
→ full verification
```

### Refactor

```text
codebase-design
→ request-refactor-plan
→ expand-contract tickets
→ implement one migration batch
→ review dependency/behavior
→ contract old form
```

### QA/report issue

```text
qa
→ reproduce and gather context
→ triage
→ needs-info / ready-for-agent / wontfix
→ to-spec only after scope is clear
```

### Perubahan besar/greenfield

```text
wayfinder
→ resolve one decision ticket per session
→ update map Decisions so far
→ graduate new tickets from fog
→ to-spec
→ to-tickets
→ implement
```

---

## 4. Alur bisnis SIMTAS yang harus dijaga

### Proposal dan Tugas Akhir

1. Mahasiswa mengajukan Proposal.
2. Sistem memvalidasi satu Tugas Akhir aktif per Mahasiswa.
3. Kaprodi meninjau dan menyetujui/menolak.
4. Setelah disetujui, Kaprodi menetapkan Dosen Pembimbing.
5. Setiap transition dicatat di audit log.

### Bimbingan dan dokumen

1. Mahasiswa mencatat konsultasi dan mengunggah dokumen sesuai type/version.
2. Dosen Pembimbing meninjau dan memberi keputusan atau Revisi.
3. Mahasiswa mengunggah revisi baru.
4. Status dokumen dan feedback harus terlihat oleh actor yang berhak.

### Seminar

1. Mahasiswa hanya dapat mengajukan Seminar jika eligibility terpenuhi.
2. Kaprodi memvalidasi dan menjadwalkan.
3. Dosen Pembimbing dan Dosen Penguji menerima assignment/notification.
4. Penilaian memakai fixed-weight rubric.
5. Hasil menghasilkan keputusan dan Revisi bila diperlukan.

### Sidang dan Surat Tugas

1. Sistem memeriksa semua prerequisite Seminar, dokumen, dan Revisi.
2. Kaprodi menyetujui Surat Tugas bila policy terpenuhi.
3. Sidang dijadwalkan dengan actor yang tepat.
4. Dosen Penguji mengisi scoring dan remarks.
5. Status akhir menentukan kelayakan graduation/clearance.

### Arsip dan reporting

1. Tugas Akhir yang selesai memiliki final document dan metadata lengkap.
2. Arsip tidak dapat diubah oleh actor biasa.
3. Dashboard membaca data melalui query yang tetap mengikuti scope dan deleted/archived policy.
4. Audit log dan email log dapat dicari oleh Admin Fakultas sesuai authorization.

---

## 5. Definition of Ready

Ticket boleh diberi label `ready-for-agent` jika semua ini terjawab:

- [ ] actor dan outcome jelas;
- [ ] istilah domain konsisten dengan `CONTEXT.md`;
- [ ] acceptance criteria dapat diuji;
- [ ] state transition dan forbidden transition jelas;
- [ ] API/UI contract cukup spesifik;
- [ ] seam test sudah disepakati;
- [ ] dependency/blocking ticket ditulis;
- [ ] migration dan rollback impact diketahui;
- [ ] out-of-scope ditulis;
- [ ] tidak ada keputusan domain penting yang masih tersembunyi.

## 6. Definition of Done

Ticket/PR hanya selesai jika:

- [ ] behavior acceptance criteria bekerja;
- [ ] authorization dan ownership test ada;
- [ ] validation/error/loading/empty state ditangani;
- [ ] audit/notification side effect sesuai contract;
- [ ] migration aman dijalankan dan tidak merusak existing data;
- [ ] test ditulis pada public seam;
- [ ] targeted test, typecheck, dan suite relevan lulus;
- [ ] `git diff --check` bersih;
- [ ] code review Standards dan Spec selesai;
- [ ] dokumentasi/CONTEXT/ADR diperbarui bila ada keputusan durable;
- [ ] PR tidak berisi secret atau data sensitif;
- [ ] deployment smoke test dan rollback plan tersedia bila production-impacting.

---

## 7. Incident flow

### Severity

- **SEV-1:** data corruption, authentication bypass, production down, atau seluruh flow utama tidak dapat dipakai.
- **SEV-2:** role/flow utama terganggu tetapi ada workaround terbatas.
- **SEV-3:** bug minor, UI issue, atau non-critical reporting issue.

### Tindakan

1. `diagnosing-bugs` untuk reproduce dan membatasi blast radius.
2. Jika security/data loss, hentikan deployment dan cabut access yang terdampak.
3. Buat incident issue dengan timestamp, symptom, impact, dan evidence.
4. Pilih mitigation paling reversible: feature flag/config/rollback.
5. Lindungi audit log dan bukti; jangan menghapus data untuk menyembunyikan failure.
6. Tambahkan regression test setelah behavior stabil.
7. Review root cause dan buat ADR hanya bila keputusan arsitektural berubah.
8. Tutup incident dengan timeline, impact, fix, verification, dan follow-up ticket.

---

## 8. Handoff antar-session

Gunakan `handoff` bila context window tidak cukup. Handoff wajib berisi:

- destination/issue/map;
- current branch dan commit;
- keputusan yang sudah dibuat;
- file/module yang diperiksa;
- test yang sudah dijalankan dan hasilnya;
- blocker/ambiguity;
- next exact action;
- perubahan yang belum di-commit.

Session berikutnya wajib memverifikasi repository state sebelum melanjutkan, bukan mempercayai handoff tanpa cek.

---

## 9. Checklist maintainer harian

```text
[ ] Apakah perubahan ini punya outcome domain yang jelas?
[ ] Apakah actor dan authorization sudah eksplisit?
[ ] Apakah state transition dan edge case sudah diuji?
[ ] Apakah ticket ini vertical slice, bukan horizontal layer?
[ ] Apakah migration dan rollback aman?
[ ] Apakah audit log dan notification behavior dipertimbangkan?
[ ] Apakah test memakai public seam?
[ ] Apakah CI green dan code review selesai?
[ ] Apakah production smoke test dilakukan?
[ ] Apakah dokumentasi/domain vocabulary tetap konsisten?
```
