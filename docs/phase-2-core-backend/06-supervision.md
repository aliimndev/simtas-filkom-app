# Job 06 — Supervision & Log Konsultasi

**Phase:** 2 — Core Backend
**Referensi PRD:** Section 6.3 (FR-CONSULT-001 s/d FR-CONSULT-003)
**Prerequisites:** Job 05 (Thesis Submission) ✅
**Estimasi:** 2 hari

---

## Objective

Implementasi modul manajemen bimbingan: pencatatan log konsultasi oleh mahasiswa atau dosen pembimbing, approval log oleh dosen, dan tracking counter bimbingan. Setelah job ini selesai, riwayat bimbingan dapat dicatat dan dipantau secara lengkap.

---

## Checklist

### Consultation Repository & Use Case
- [x] Buat `backend/internal/domain/repository/consultation_repository.go` — interface:
  ```go
  type ConsultationRepository interface {
    Create(ctx context.Context, log *entity.ConsultationLog) error
    FindByID(ctx context.Context, id uuid.UUID) (*entity.ConsultationLog, error)
    FindByThesisID(ctx context.Context, thesisID uuid.UUID, filter ConsultationFilter) ([]*entity.ConsultationLog, int64, error)
    Update(ctx context.Context, log *entity.ConsultationLog) error
    Approve(ctx context.Context, id uuid.UUID, approvedBy uuid.UUID) error
    CountApprovedByThesisID(ctx context.Context, thesisID uuid.UUID) (int, error)
  }
  ```
- [x] `ConsultationFilter`: `Status`, `DateFrom`, `DateTo`, `Page`, `PerPage`
- [x] Buat `backend/internal/usecase/consultation_usecase.go`

### Validasi Akses (Ownership)
Buat helper `backend/internal/usecase/helpers.go` — fungsi yang akan digunakan berulang:
- [x] `IsThesisOwner(userID, thesisID)` — cek apakah user adalah mahasiswa pemilik thesis
- [x] `IsSupervisor(userID, thesisID)` — cek apakah user adalah pembimbing thesis ini
- [x] `IsExaminer(userID, seminarID/defenseID)` — cek apakah user adalah penguji (digunakan di job berikutnya)

### Handler — Consultation Log

**POST `/api/v1/theses/:thesis_id/consultations`** _(Mahasiswa pemilik + Dosen Pembimbing thesis ini)_
- [x] Request body:
  ```json
  {
    "consultation_date": "2026-10-15",
    "topics_discussed": "Pembahasan BAB 2 tinjauan pustaka",
    "notes": "Perlu tambahkan referensi terbaru",
    "follow_up": "Upload draft BAB 2 revisi minggu depan",
    "attachment_url": null
  }
  ```
- [x] Validasi:
  - `consultation_date` tidak boleh di masa depan
  - `topics_discussed` tidak boleh kosong
  - Thesis harus berstatus `in_progress` atau lebih lanjut
  - User harus pemilik thesis atau dosen pembimbingnya
- [x] Status awal: `pending`
- [x] Audit log: `CONSULTATION_CREATED`
- [x] Email notification (stub):
  - Jika dibuat oleh mahasiswa → email ke dosen pembimbing
  - Jika dibuat oleh dosen → email ke mahasiswa

**GET `/api/v1/theses/:thesis_id/consultations`** _(Mahasiswa pemilik + Dosen Pembimbing + Admin + Kaprodi)_
- [x] Query params: `status`, `date_from`, `date_to`, `page`, `per_page`
- [x] Return list log konsultasi diurutkan dari terbaru
- [x] Include summary di response:
  ```json
  {
    "success": true,
    "data": {
      "consultations": [ { log objects } ],
      "summary": {
        "total": 12,
        "approved": 10,
        "pending": 2,
        "last_consultation_date": "2026-10-15"
      }
    },
    "meta": { "page": 1, "per_page": 20, "total": 12 }
  }
  ```

**GET `/api/v1/theses/:thesis_id/consultations/:id`** _(akses sama)_
- [x] Return detail 1 log konsultasi

**PUT `/api/v1/theses/:thesis_id/consultations/:id`** _(Pembuat log, selama masih `pending`)_
- [x] Hanya bisa edit jika status masih `pending`
- [x] Field yang bisa diupdate: `topics_discussed`, `notes`, `follow_up`, `attachment_url`, `consultation_date`
- [x] Audit log: `CONSULTATION_UPDATED`

**PATCH `/api/v1/theses/:thesis_id/consultations/:id/approve`** _(Dosen Pembimbing thesis ini only)_
- [x] Validasi: status harus `pending`
- [x] Set status → `approved`, set `approved_by`, `approved_at`
- [x] Increment counter bimbingan (diambil dari `CountApprovedByThesisID`)
- [x] Audit log: `CONSULTATION_APPROVED`
- [x] Email notification ke mahasiswa

**DELETE `/api/v1/theses/:thesis_id/consultations/:id`** _(Pembuat log, selama masih `pending`)_
- [x] Soft delete atau hard delete (log yang belum diapprove)
- [x] Validasi: tidak bisa hapus jika sudah `approved`
- [x] Audit log: `CONSULTATION_DELETED`

### Summary Endpoint

**GET `/api/v1/theses/:thesis_id/consultations/summary`** _(sama dengan akses GET list)_
- [x] Return ringkasan tanpa pagination:
  ```json
  {
    "total_consultations": 12,
    "approved_count": 10,
    "pending_count": 2,
    "last_consultation_date": "2026-10-15",
    "average_interval_days": 7,
    "consultations_this_month": 3
  }
  ```

### Attachment Upload (Opsional — bisa di-stub dulu)
- [x] Jika `attachment_url` disertakan, validasi URL valid
- [x] Upload file attachment ke storage diimplementasikan di Job 21 (Storage Integration)
- [x] Untuk saat ini, field `attachment_url` menerima URL langsung (manual atau pre-uploaded)

---

## Done Criteria

- [x] `POST /api/v1/theses/:id/consultations` oleh mahasiswa pemilik → log dibuat, status `pending`
- [x] `POST /api/v1/theses/:id/consultations` oleh mahasiswa lain → `403 Forbidden`
- [x] `POST /api/v1/theses/:id/consultations` dengan tanggal masa depan → `400 Bad Request`
- [x] `PATCH .../consultations/:id/approve` oleh dosen pembimbing → status jadi `approved`
- [x] `PATCH .../consultations/:id/approve` oleh dosen yang bukan pembimbing thesis ini → `403`
- [x] `PATCH .../consultations/:id/approve` jika status sudah `approved` → `422`
- [x] `GET .../consultations` → return list dengan summary (total, approved, pending)
- [x] `PUT .../consultations/:id` setelah status `approved` → `422 Unprocessable Entity`
- [x] `CountApprovedByThesisID` mengembalikan angka yang benar setelah beberapa approval
- [x] Email notification (stub) tercatat di console
- [x] Semua action tercatat di `audit_logs`
