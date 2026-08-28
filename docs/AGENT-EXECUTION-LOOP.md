# SIMTAS FILKOM — Agent Execution Loop

Dokumen ini adalah protocol eksekusi untuk AI agent. `docs/ENGINEERING-WORKFLOW.md` menjelaskan prinsip dan lifecycle; file ini menjelaskan bagaimana agent menjalankan pekerjaan secara berurutan dan melakukan self-correction.

## 1. Cara menjalankan

### Shorthand maintainer: `continue job`

Jika maintainer hanya mengirim:

```text
continue job
```

agent harus otomatis:

1. membaca `AGENTS.md`, `CONTEXT.md`, `PRODUCT.md`, `docs/ENGINEERING-WORKFLOW.md`, dan dokumen ini;
2. menemukan map Wayfinder terbuka untuk pekerjaan SIMTAS;
3. mencari child issue frontier pertama yang berlabel `ready-for-agent`, tidak terblokir, dan belum di-claim;
4. jika frontier memerlukan keputusan manusia, menanyakan **satu pertanyaan keputusan saja**;
5. jika frontier siap dikerjakan, meng-claim dan menjalankan execution loop sampai selesai atau safety stop;
6. menyimpan evidence dan status di GitHub Issue agar invocation berikutnya dapat melanjutkan tanpa konteks manual.

Agent tidak boleh meminta maintainer mengulang nomor issue, scope, atau rencana yang sudah ada di map/issue. Agent hanya meminta input jika memang ada keputusan HITL, akses, secret, atau safety stop yang belum tersedia.

Jika tidak ada map terbuka, agent mencari issue `ready-for-agent` paling relevan berdasarkan dependency dan roadmap. Jika tidak ada ticket yang siap, agent melaporkan blocker dan tidak mengarang pekerjaan baru.

### Invocation lengkap (opsional)

AI agent tidak berjalan sebagai daemon otomatis setelah satu jawaban. Jalankan setiap session dengan instruksi eksplisit berikut:

```text
Jalankan Agent Execution Loop untuk SIMTAS.
Destination/map: <URL atau nomor issue map>
Mode: implement satu frontier ticket, lalu berhenti setelah evidence lengkap.

Baca AGENTS.md, CONTEXT.md, PRODUCT.md,
docs/ENGINEERING-WORKFLOW.md, dan docs/AGENT-EXECUTION-LOOP.md.
Gunakan skill Matt Pocock yang sesuai. Jangan mengambil ticket yang blocked.
```

Untuk melanjutkan session berikutnya:

```text
Lanjutkan Agent Execution Loop SIMTAS dari frontier ticket map <URL atau nomor>.
Baca ulang issue map, status git, dan evidence terakhir sebelum bekerja.
```

Satu session default hanya mengerjakan **satu ticket frontier**. Ini menjaga context, review, dan rollback tetap dapat dilacak. Research ticket yang independen boleh dikerjakan paralel jika hasilnya tidak mengubah source code production.

## 2. Source of truth dan status

Urutan source of truth:

1. domain vocabulary: `CONTEXT.md` dan `PRODUCT.md`;
2. keputusan durable: `docs/adr/`;
3. destination/map dan child issues di GitHub;
4. acceptance criteria ticket;
5. repository state dan test evidence;
6. dokumen workflow.

Status kerja berada di GitHub Issues:

- `needs-triage`: belum siap dikerjakan;
- `needs-info`: menunggu jawaban manusia;
- `ready-for-agent`: acceptance criteria lengkap;
- `ready-for-human`: membutuhkan keputusan/akses manusia;
- `wontfix`: tidak dikerjakan.

Ticket hanya boleh dipilih jika:

- berlabel `ready-for-agent`;
- tidak memiliki blocker terbuka;
- belum memiliki assignee;
- acceptance criteria dan seam test cukup jelas.

Agent wajib claim ticket sebelum mengubah code:

```bash
gh issue edit <ticket> --add-assignee @me
```

Jangan memakai urutan nomor issue sebagai pengganti dependency. Frontier ditentukan oleh blocker yang sudah selesai.

## 3. State machine execution

```text
DISCOVER
  → CLAIM
  → UNDERSTAND
  → PLAN_SEAM
  → RED
  → GREEN
  → VERIFY
  → REVIEW
  → FIX_REVIEW_FINDINGS
  → FINAL_VERIFY
  → RECORD_EVIDENCE
  → NEXT_FRONTIER
```

Transition ke `BLOCKED` bila dependency, akses, keputusan domain, atau environment belum tersedia. Transition ke `ESCALATE` bila safety stop terpenuhi. Agent tidak boleh melompati `FINAL_VERIFY`.

## 4. Prosedur setiap ticket

### DISCOVER

1. Baca issue lengkap beserta komentar dan labels.
2. Baca map dan semua blocker yang relevan.
3. Baca `CONTEXT.md`, `PRODUCT.md`, ADR area terkait, dan workflow.
4. Cek `git status`, branch, commit terakhir, dan CI terbaru.
5. Cari implementasi yang sudah ada berdasarkan konsep domain.

Jika behavior ternyata sudah ada, jangan menduplikasi. Laporkan lokasi implementasinya dan usulkan menutup ticket sebagai already implemented/wontfix bila sesuai.

### CLAIM

Assign issue kepada agent. Jangan mengklaim ticket jika tidak dapat menyelesaikannya dalam satu session atau jika ada keputusan HITL yang belum dijawab.

### UNDERSTAND

Tuliskan secara internal atau di plan:

- actor dan outcome;
- state sebelum/sesudah;
- allowed dan forbidden transitions;
- authorization/ownership;
- error dan duplicate behavior;
- audit/notification side effects;
- migration dan rollback impact;
- out of scope.

Jika istilah domain bertentangan dengan `CONTEXT.md`, berhenti dan gunakan `domain-modeling` atau `grilling`.

### PLAN_SEAM

Tentukan seam publik tertinggi yang akan diuji. Contoh SIMTAS:

- Hono API request/response untuk flow backend;
- UI interaction untuk behavior halaman;
- health/smoke endpoint untuk deployment.

Tulis daftar behavior test sebelum implementasi. Jangan test private function atau mock internal collaborator tanpa alasan kuat.

### RED

Untuk setiap behavior:

1. tulis satu test pada seam yang disepakati;
2. jalankan test target;
3. pastikan test gagal karena behavior belum tersedia;
4. simpan output failure sebagai bukti red.

Jika test gagal karena setup/fixture/environment, perbaiki setup dulu atau escalate; jangan menganggapnya red yang valid.

### GREEN

Implementasi minimal untuk satu tracer bullet end-to-end:

```text
schema/migration bila diperlukan
  → domain/service rule
  → Hono route + authorization
  → audit/notification side effect
  → Svelte UI behavior bila ticket mencakup UI
  → integration/UI test
```

Jangan menambahkan abstraction atau fitur yang belum diminta. Jalankan test target setelah setiap perubahan bermakna.

### VERIFY

Jalankan check yang relevan. Minimum untuk perubahan TypeScript/API/UI:

```bash
git diff --check
bun run typecheck
bun run check
bun run test
bun run build
```

Untuk API/database, pastikan PostgreSQL test tersedia dan jalankan integration tests. Untuk UI, jalankan test route/component yang berubah dan build web. Untuk migration, jalankan migration dari database bersih dan database dengan fixture.

### SELF-CORRECTION LOOP

Jika verification gagal:

1. catat command, exit code, file/error, dan klasifikasi failure;
2. cari root cause, bukan sekadar memperlebar assertion;
3. ubah code/test/config seminimal mungkin;
4. jalankan ulang check yang gagal;
5. jalankan kembali check terkait;
6. ulangi sampai pass.

Klasifikasi failure:

- **code defect:** agent wajib memperbaiki;
- **test defect:** perbaiki test hanya jika expected behavior memang benar;
- **contract defect:** kembali ke spec/domain decision;
- **environment defect:** perbaiki setup jika aman, jika tidak escalate;
- **unrelated baseline defect:** jangan menyamarkan; catat evidence dan scope.

Batas aman:

- ulangi failure yang sama maksimal 3 siklus tanpa perubahan diagnosis;
- jika setelah 3 siklus tetap sama, berhenti dengan status `ESCALATE`;
- jangan menghapus test, menurunkan assertion, skip test, atau memakai `continue-on-error` untuk menyembunyikan failure;
- jangan mengubah migration production secara destructive untuk membuat test pass.

### REVIEW

Setelah test lokal pass:

1. baca diff lengkap;
2. jalankan `code-review` pada fixed point branch;
3. review dua axis terpisah: Standards dan Spec;
4. cek authorization, ownership, state transition, side effect, migration, dan UI states;
5. jalankan `web-design-guidelines` atau `review` bila ticket menyentuh UI.

### FIX_REVIEW_FINDINGS

Semua finding blocker harus diperbaiki. Untuk setiap finding:

```text
finding → diagnosis → focused edit → targeted test → full relevant verification
```

Finding yang merupakan trade-off nyata harus dicatat pada issue/ADR, bukan diam-diam diabaikan.

### FINAL_VERIFY

Sebelum menyatakan ticket selesai, ulangi verification dari keadaan repository terbaru. Evidence harus fresh dan mencantumkan:

- command yang dijalankan;
- exit code;
- hasil test/build;
- migration check jika ada;
- review result;
- `git diff --check`;
- file yang berubah.

Gunakan `verification-before-completion`. Jangan memakai hasil command dari session lama sebagai bukti final.

### RECORD_EVIDENCE

Update issue dengan komentar ringkas:

```markdown
## Implementation evidence

- Behavior delivered: ...
- Tests: `<command>` — pass/fail summary
- Build/typecheck: `<command>` — pass/fail summary
- Review: Standards ...; Spec ...
- Migration/rollback: ...
- Deployment/smoke test: ...
- Follow-up: ...
```

Issue hanya boleh ditutup jika acceptance criteria tercentang dan evidence final tersedia. Jika belum, biarkan open dengan status/blocker yang benar.

## 5. Sequential module plan SIMTAS

Urutan default dari `docs/ROADMAP.md`:

```text
users
→ roles
→ academic_years
→ theses / Proposal
→ title_change_requests
→ consultation_logs
→ seminars / Seminar
→ thesis_defenses / Sidang
→ documents
→ thesis_archives / Arsip
→ notifications
→ email_logs
→ audit_logs
→ dashboard
```

Untuk gap saat ini, jangan melompat langsung ke implementasi Seminar/Sidang/Arsip sebelum canonical status dan compatibility decision dikunci. Setelah keputusan itu selesai, frontier yang direkomendasikan:

1. Seminar submission + list/detail;
2. Seminar scheduling + Penguji;
3. Seminar scoring + result;
4. Sidang submission + list/detail;
5. Sidang scheduling + Penguji;
6. Sidang scoring + Revisi + graduation;
7. Arsip + storage/download;
8. upcoming schedules + dashboard.

Satu ticket harus tetap vertical slice, bukan membuat seluruh service backend dulu lalu UI belakangan.

## 6. Safety stops

Agent wajib berhenti dan meminta persetujuan manusia jika:

- harus memilih policy domain yang belum diputuskan;
- akan menghapus data atau melakukan destructive migration;
- membutuhkan secret, credential, akses VPS, atau production database;
- akan melakukan deployment production, rollback production, atau mengubah infrastructure;
- menemukan kemungkinan security vulnerability atau data exposure;
- test baseline gagal dan tidak jelas apakah perubahan yang menyebabkan;
- scope ticket berbeda dari acceptance criteria;
- harus membuat banyak ticket baru karena destination belum jelas.

Jangan memaksa loop sampai green dengan mengorbankan integritas domain, data, atau security.

## 7. Apa yang otomatis dan apa yang tidak

Protocol ini membuat agent mampu mengikuti loop secara konsisten, tetapi file Markdown sendiri tidak dapat memanggil AI secara otomatis. GitHub Actions hanya menjalankan command yang sudah dikonfigurasi; ia tidak boleh diberi akses menjalankan agent production tanpa sandbox, approval, dan secret policy.

Yang dilakukan agent pada setiap invocation:

1. membaca map dan frontier;
2. mengerjakan satu ticket;
3. memperbaiki error berdasarkan evidence;
4. berhenti setelah ticket selesai atau safety stop;
5. session berikutnya melanjutkan dari issue/map yang sama.

Dengan demikian urutan dan progress tersimpan di GitHub Issues, bukan hanya di memory satu conversation.
