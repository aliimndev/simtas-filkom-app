# ADR-001: Title Change Request — Perubahan Judul Pasca-Pembimbing

**Status:** Proposed  
**Date:** 2026-08-06  
**Phase:** 2 — Core Backend (diusulkan setelah Job 05: Thesis Submission)  
**Prerequisites:** Job 05 (Thesis Submission + Supervisor Assignment) ✅, Job 11 (Email Notification) ✅  
**Referensi PRD:** Section 6.2 (FR-THESIS-*), Section 6.10 (FR-NOTIF-001)

---

## Context

Saat ini SIMTAS FILKOM hanya mengizinkan perubahan judul pada tahap pengajuan awal (sebelum disetujui Kaprodi). Setelah pembimbing ditetapkan dan thesis berstatus `approved` atau `in_progress`, mahasiswa tidak memiliki cara formal untuk mengajukan perubahan judul di dalam sistem.

Berdasarkan arahan dosen, perubahan judul setelah pembimbing ditetapkan harus:
1. Dilakukan melalui Dosen Pembimbing (bukan Kaprodi).
2. Memiliki histori/permintaan persetujuan yang ter-record.
3. Tetap konsisten dengan alur thesis yang sudah ada.

Keputusan ini dibuat untuk memenuhi kebutuhan akademik Fakultas Ilmu Komputer tanpa mengubah alur existing yang sudah berjalan.

---

## Decision

Kami akan menambahkan fitur **Title Change Request** dengan komponen berikut:

### 1. Business Flow

```
Mahasiswa
  |
  | Ajukan perubahan judul (hanya jika thesis APPROVED/IN_PROGRESS + ada pembimbing)
  ↓
title_change_requests (status: PENDING)
  |
  | Dosen Pembimbing review
  ├── APPROVED → update theses.title + audit_logs + email ke mahasiswa
  ├── REJECTED → judul tetap, audit_logs + email ke mahasiswa
  └── CANCELLED → mahasiswa tarik permintaan sebelum diproses
```

### 2. Actor & Permission Matrix

| Actor | Create Request | Approve/Reject | Cancel | Read (all requests) |
|-------|---------------|----------------|--------|---------------------|
| Mahasiswa | ✅ (pemilik thesis) | ❌ | ✅ (PENDING miliknya) | ✅ (miliknya) |
| Dosen Pembimbing | ❌ | ✅ (assigned thesis, PENDING only) | ❌ | ✅ (mahasiswa bimbingan) |
| Kaprodi | ❌ | ❌ | ❌ | ✅ (read-only, monitoring) |
| Admin | ❌ | ❌ | ❌ | ✅ (read-only, audit) |

### 3. State Machine

```
PENDING
  ├─► APPROVED  (oleh Dosen Pembimbing)
  ├─► REJECTED  (oleh Dosen Pembimbing)
  └─► CANCELLED (oleh Mahasiswa, sebelum di-review)
```

**Rules:**
- Semua transisi hanya berlaku dari `PENDING`.
- `APPROVED`/`REJECTED`/`CANCELLED` adalah terminal states (immutable).
- Tidak ada cooldown atau limit pengajuan ulang jika `REJECTED`.
- Atomik: saat `APPROVED`, sistem langsung update `theses.title` dan catat audit log dalam satu transaksi database.


### 5. API Contract

Base path: `/api/v1`

#### 5.1 Create Request
**POST** `/theses/{thesis_id}/title-change-requests`  
**Auth:** Mahasiswa (pemilik thesis)  
**Request:**
```json
{
  "requested_title": "Judul Baru Skripsi",
  "reason": "Alasan perubahan (opsional)"
}
```
**Response 201:**
```json
{
  "id": "uuid",
  "thesis_id": "uuid",
  "previous_title": "Judul Lama",
  "requested_title": "Judul Baru",
  "status": "PENDING",
  "requested_by": "uuid",
  "reviewed_by": null,
  "review_notes": null,
  "cancelled_by": null,
  "cancelled_at": null,
  "created_at": "2026-08-06T...",
  "updated_at": "2026-08-06T..."
}
```
**Error Cases:**
- `400` — Thesis belum ada pembimbing aktif
- `400` — Thesis status bukan `approved` atau `in_progress`
- `400` — Sudah ada request `PENDING` untuk thesis ini
- `403` — Bukan pemilik thesis
- `404` — Thesis tidak ditemukan

#### 5.2 List Requests (for a thesis)
**GET** `/theses/{thesis_id}/title-change-requests`  
**Auth:** Mahasiswa (pemilik), Dosen Pembimbing (assigned), Kaprodi, Admin  
**Response 200:**
```json
{
  "data": [
    {
      "id": "uuid",
      "previous_title": "...",
      "requested_title": "...",
      "status": "PENDING",
      "requested_by": { "full_name": "Mahasiswa", "nim_nidn": "123" },
      "reviewed_by": { "full_name": "Dosen", "nim_nidn": "456" } | null,
      "review_notes": null,
      "created_at": "..."
    }
  ]
}
```
#### 5.3 Approve Request
**POST** `/title-change-requests/{id}/approve`  
**Auth:** Dosen Pembimbing (assigned)  
**Request:**
```json
{
  "review_notes": "Catatan persetujuan (opsional)"
}
```
**Response 200:** Updated request object  
**Side Effects:**
- `theses.title = requested_title`
- `audit_logs`: `TITLE_CHANGE_APPROVED`, `THESIS_TITLE_UPDATED`
- Email ke mahasiswa

**Error Cases:**
- `400` — Status bukan `PENDING`
- `403` — Bukan dosen pembimbing yang ditugaskan
- `404` — Request tidak ditemukan

#### 5.4 Reject Request
**POST** `/title-change-requests/{id}/reject`  
**Auth:** Dosen Pembimbing (assigned)  
**Request:**
```json
{
  "review_notes": "Alasan penolakan (wajib)"
}
```
**Response 200:** Updated request object  
**Side Effects:**
- `audit_logs`: `TITLE_CHANGE_REJECTED`
- Email ke mahasiswa

**Error Cases:**
- `400` — Status bukan `PENDING`
- `400` — `review_notes` kosong
- `403` — Bukan dosen pembimbing yang ditugaskan
- `404` — Request tidak ditemukan

#### 5.5 Cancel Request
**POST** `/title-change-requests/{id}/cancel`  
**Auth:** Mahasiswa (pemilik)  
**Request:** `{}` (empty)  
**Response 200:** Updated request object  
**Side Effects:**
- `audit_logs`: `TITLE_CHANGE_CANCELLED`
- Email ke Dosen Pembimbing (opsional, untuk transparansi)

**Error Cases:**
- `400` — Status bukan `PENDING`
- `403` — Bukan pemilik request
- `404` — Request tidak ditemukan

### 6. Frontend Flow

#### 6.1 Mahasiswa — Halaman Detail Thesis (`/thesis`)
- Tampilkan section **"Riwayat Perubahan Judul"** di bawah detail thesis.
- Tombol **"Ajukan Perubahan Judul"** hanya muncul jika:
  - User adalah pemilik thesis.
  - Thesis memiliki `supervisors` aktif.
  - Thesis status = `approved` atau `in_progress`.
  - Tidak ada `title_change_requests` dengan status `PENDING`.
- Form perubahan judul (modal atau page):
  - Judul Saat Ini (readonly)
  - Judul Baru (required, max 500 chars, min 10 kata)
  - Alasan Perubahan (optional)
- Setelah submit: tampilkan toast, refresh data, mahasiswa melihat status di detail thesis dan dashboard.

**Section Riwayat Perubahan Judul (read-only table/card):**
| Field | Source |
|-------|--------|
| Judul Lama | `previous_title` |
| Judul Baru | `requested_title` |
| Status | `status` |
| Reviewer | `reviewed_by.full_name` |
| Tanggal Pengajuan | `created_at` |
| Catatan Pembimbing | `review_notes` |

#### 6.2 Dosen Pembimbing — Dashboard
- Tambahkan menu item: **"Review Perubahan Judul"** di dashboard dosen.
- Halaman menampilkan daftar `title_change_requests` dengan status `PENDING` untuk mahasiswa bimbingan.
- Setiap item menampilkan:
  - Nama mahasiswa + NIM
  - Judul lama → judul baru
  - Alasan perubahan
  - Tombol **Review** → buka modal konfirmasi

**Modal Konfirmasi Review:**
```
Konfirmasi Perubahan Judul

Judul Sebelumnya:
[judul lama]

Judul Baru:
[judul baru]

Alasan Mahasiswa:
[alasan]

Catatan Pembimbing:
[textarea]

[Tolak] [Setujui]
```

#### 6.3 Kaprodi & Admin
- Akses read-only via halaman manajemen thesis yang sudah ada, atau bisa ditambahkan filter khusus nanti.
- Tidak ada tombol aksi approve/reject/cancel.

### 7. Security & Validation Rules

#### 7.1 RBAC & Ownership
- Setiap endpoint memverifikasi role + ownership/assignment sebelum izinkan aksi.
- Gunakan middleware `RequireRole` + guard spesifik di use case/handler.

#### 7.2 State Transition Validation
- Approve/Reject/Create hanya valid jika status request = `PENDING`.
- Gunakan state machine pattern yang konsisten dengan `thesis_state.go` yang sudah ada.
- Jika transisi invalid → `422 Unprocessable Entity`.

#### 7.3 Thesis Status Validation
- Create request hanya diizinkan jika thesis.status ∈ {`approved`, `in_progress`}.
- Validasi dilakukan sebelum insert ke database.

#### 7.4 Supervisor Validation
- Approve/Reject hanya boleh oleh user yang ada di `thesis_supervisors` untuk thesis tersebut.
- Query supervisor assignment dari repository, jangan andal user-provided ID tanpa verifikasi.

#### 7.5 Concurrency Control
- Enforce `UNIQUE(thesis_id) WHERE status = 'PENDING'` di database.
- Di aplikasi: gunakan `INSERT ... ON CONFLICT DO NOTHING` atau check terlebih dahulu dalam transaksi.
- Jika ada race condition, kembalikan `409 Conflict` dengan message: "Sudah ada permintaan perubahan judul yang sedang diproses."

#### 7.6 Audit Logging
Setiap aksi wajib membuat record di `audit_logs`:
- `TITLE_CHANGE_REQUESTED` — saat mahasiswa submit
- `TITLE_CHANGE_APPROVED` — saat dosen approve
- `TITLE_CHANGE_REJECTED` — saat dosen reject
- `TITLE_CHANGE_CANCELLED` — saat mahasiswa cancel
- `THESIS_TITLE_UPDATED` — saat judul berhasil di-update

**Audit fields:** `actor_id`, `action`, `entity_type = 'title_change_request'`, `entity_id`, `old_value` (previous_title), `new_value` (requested_title), `ip_address`, `user_agent`.

### 8. Email Notification Events

| Event | Recipients | Trigger |
|-------|-----------|---------|
| `TITLE_CHANGE_REQUESTED` | Mahasiswa + semua Dosen Pembimbing yang assigned | Setelah `PENDING` created |
| `TITLE_CHANGE_APPROVED` | Mahasiswa | Setelah `APPROVED` + title updated |
| `TITLE_CHANGE_REJECTED` | Mahasiswa | Setelah `REJECTED` |
| `TITLE_CHANGE_CANCELLED` | Dosen Pembimbing (opsional) | Setelah `CANCELLED` |

**Email template** akan mengikuti pola yang sudah ada di `backend/pkg/email/templates/`.

### 9. Migration Strategy

1. **Backend migration:** Tambahkan tabel `title_change_requests` + unique index partial untuk `PENDING`.
2. **Repository + Usecase:** Implementasi baru tanpa mengubah existing thesis flow.
3. **API endpoints:** Ditambahkan di router yang sudah ada (`backend/internal/delivery/http/`).
4. **Frontend:** Ditambahkan di halaman thesis detail + dashboard dosen.
5. **Rollback:** Jika ada issue, bisa disable endpoint tanpa mempengaruhi thesis flow utama.

### 10. Alternatives Considered

- **Opsi 1: Single table `thesis_versions`** — ditolak karena terlalu kompleks untuk kebutuhan saat ini. Audit trail sudah tercakup oleh `audit_logs`.
- **Opsi 3: Kaprodi juga approve perubahan judul** — ditolak sesuai arahan dosen, perubahan judul pasca-pembimbing hanya di Dosen Pembimbing.
- **Opsi C: Izinkan perubahan di semua status thesis** — ditolak karena perubahan judul sebaiknya tidak diizinkan setelah seminar/sidang.

---

## References

- PRD Section 6.2 — Thesis Submission & Review
- PRD Section 6.10 — Notification System
- `docs/pertanyaan-kaprodi-flow.md` — Context Kaprodi review flow

