# Job 07 — Document Management (Upload & Approval Workflow)

**Phase:** 2 — Core Backend
**Referensi PRD:** Section 6.4 (FR-DOC-001 s/d FR-DOC-004)
**Prerequisites:** Job 06 (Supervision) ✅
**Estimasi:** 3 hari

---

## Objective

Implementasi sistem manajemen dokumen dengan upload bertahap dan mekanisme approval: mahasiswa upload dokumen, dosen pembimbing review dan approve/minta revisi, dengan gate logic yang memastikan mahasiswa tidak bisa lanjut ke tahap berikutnya sebelum dokumen prerequisite disetujui. Versioning sederhana (riwayat upload) juga diimplementasikan.

---

## Checklist

### Document Repository & Use Case
- [ ] Buat `backend/internal/domain/repository/document_repository.go` — interface:
  ```go
  type DocumentRepository interface {
    Create(ctx context.Context, doc *entity.Document) error
    FindByID(ctx context.Context, id uuid.UUID) (*entity.Document, error)
    FindByThesisID(ctx context.Context, thesisID uuid.UUID, filter DocumentFilter) ([]*entity.Document, int64, error)
    FindLatestByType(ctx context.Context, thesisID uuid.UUID, docType string, chapterNum *int) (*entity.Document, error)
    UpdateStatus(ctx context.Context, id uuid.UUID, status string, reviewerID uuid.UUID, notes string) error
    GetVersionHistory(ctx context.Context, thesisID uuid.UUID, docType string, chapterNum *int) ([]*entity.Document, error)
    IsDocumentApproved(ctx context.Context, thesisID uuid.UUID, docType string) (bool, error)
  }
  ```
- [ ] `DocumentFilter`: `DocumentType`, `Status`, `Page`, `PerPage`
- [ ] Buat `backend/internal/usecase/document_usecase.go`

### Tipe Dokumen & Gate Logic
Definisikan konstanta dan aturan gate:

```go
// internal/domain/entity/document_types.go
const (
  DocTypeProposal          = "proposal"
  DocTypeDraftChapter      = "draft_chapter"    // chapter_number: 1-5
  DocTypeSeminarDoc        = "seminar_doc"
  DocTypeDefenseDoc        = "defense_doc"
  DocTypeFinalThesis       = "final_thesis"
  DocTypeRevisionSheet     = "revision_sheet"
  DocTypeEndorsementLetter = "endorsement_letter"
)

// Gate: dokumen yang harus approved sebelum bisa ajukan seminar
var SeminarGate = []string{DocTypeSeminarDoc}

// Gate: dokumen yang harus approved sebelum bisa ajukan sidang
var DefenseGate = []string{DocTypeDefenseDoc}
```

- [ ] Buat `backend/internal/usecase/gate_checker.go`:
  ```go
  func (uc *DocumentUseCase) CanSubmitSeminar(ctx context.Context, thesisID uuid.UUID) (bool, error)
  func (uc *DocumentUseCase) CanSubmitDefense(ctx context.Context, thesisID uuid.UUID) (bool, error)
  ```

### Handler — Document Endpoints

**POST `/api/v1/theses/:thesis_id/documents`** _(Mahasiswa pemilik only)_
- [ ] Content-Type: `multipart/form-data`
- [ ] Form fields:
  - `file` — file PDF (required)
  - `document_type` — string (required)
  - `chapter_number` — integer (required jika `document_type = draft_chapter`)
  - `notes` — string (optional)
- [ ] Validasi:
  - Format file: hanya PDF
  - Ukuran file: maksimal 10 MB
  - `document_type` harus salah satu dari konstanta yang valid
  - `chapter_number` harus 1–5 jika `document_type = draft_chapter`
  - Thesis harus berstatus `in_progress` atau lebih lanjut
  - User harus pemilik thesis
- [ ] Upload file ke Supabase Storage — path: `theses/{thesis_id}/{document_type}/v{version}_{filename}`
  - **Untuk Job 07**: gunakan stub storage yang simpan file path saja (implementasi penuh di Job 21)
- [ ] Auto-increment versi berdasarkan dokumen type yang sama
- [ ] Set status awal: `pending_review`
- [ ] Audit log: `DOCUMENT_UPLOADED`
- [ ] Email notification ke dosen pembimbing (stub)
- [ ] Response: `201 Created` dengan document object

**GET `/api/v1/theses/:thesis_id/documents`** _(Mahasiswa pemilik + Dosen Pembimbing + Dosen Penguji + Admin + Kaprodi)_
- [ ] Query params: `document_type`, `status`, `page`, `per_page`
- [ ] Return daftar dokumen aktif (versi terbaru per tipe)
- [ ] Response:
  ```json
  {
    "success": true,
    "data": [
      {
        "id": "uuid",
        "document_type": "seminar_doc",
        "chapter_number": null,
        "version": 2,
        "file_name": "seminar_doc_v2.pdf",
        "file_size": 1048576,
        "status": "approved",
        "reviewer": { "id": "...", "full_name": "Dr. Ahmad" },
        "reviewer_notes": "Dokumen sudah baik",
        "reviewed_at": "2026-10-20T10:00:00Z",
        "created_at": "2026-10-18T08:00:00Z"
      }
    ],
    "meta": { ... }
  }
  ```

**GET `/api/v1/theses/:thesis_id/documents/:id`** _(akses sama)_
- [ ] Return detail dokumen termasuk info versi

**GET `/api/v1/theses/:thesis_id/documents/:id/download`** _(akses sama)_
- [ ] Generate presigned URL dari storage (expired 15 menit)
- [ ] Log download di audit_logs: `DOCUMENT_DOWNLOADED`
- [ ] Response: `{ "success": true, "data": { "download_url": "...", "expires_in": 900 } }`

**GET `/api/v1/theses/:thesis_id/documents/history`** _(akses sama)_
- [ ] Query params: `document_type`, `chapter_number`
- [ ] Return semua versi dokumen untuk tipe tertentu (riwayat upload), diurutkan dari versi terbaru
- [ ] Berguna untuk melihat revisi history

**PATCH `/api/v1/documents/:id/review`** _(Dosen Pembimbing thesis terkait only)_
- [ ] Request body:
  ```json
  {
    "decision": "approved",
    "notes": "Dokumen sudah memenuhi standar"
  }
  ```
  atau
  ```json
  {
    "decision": "revision_required",
    "notes": "Harap perbaiki BAB 2 sub-bab 2.3, referensi kurang update"
  }
  ```
- [ ] Validasi:
  - `decision`: `approved` atau `revision_required`
  - Status dokumen harus `pending_review`
  - User harus pembimbing thesis ini
- [ ] Update status dokumen
- [ ] Set `reviewer_id`, `reviewer_notes`, `reviewed_at`
- [ ] Audit log: `DOCUMENT_APPROVED` atau `DOCUMENT_REVISION_REQUESTED`
- [ ] Email notification ke mahasiswa (stub)

### Storage Interface (Stub untuk Job 07)
- [ ] Buat `backend/internal/domain/service/storage_service.go`:
  ```go
  type StorageService interface {
    Upload(ctx context.Context, path string, file io.Reader, size int64, contentType string) (string, error)
    GeneratePresignedURL(ctx context.Context, path string, expirySeconds int) (string, error)
    Delete(ctx context.Context, path string) error
  }
  ```
- [ ] Buat `backend/pkg/storage/stub_storage_service.go` — simpan file ke folder lokal `./tmp/uploads/` untuk development, return path lokal sebagai URL

### Validasi File Helper
- [ ] Buat `backend/pkg/utils/file_validator.go`:
  ```go
  func ValidatePDF(file multipart.File, header *multipart.FileHeader) error
  // Cek: ekstensi .pdf, MIME type application/pdf, ukuran <= maxSize
  func GetFileSizeInMB(size int64) float64
  ```

---

## Done Criteria

- [ ] `POST .../documents` upload PDF valid → dokumen tersimpan, status `pending_review`
- [ ] `POST .../documents` upload file non-PDF → `400 Bad Request`
- [ ] `POST .../documents` upload file >10 MB → `400 Bad Request`
- [ ] `POST .../documents` upload ke thesis orang lain → `403 Forbidden`
- [ ] Upload 2x dokumen tipe `seminar_doc` → versi ke-2 terbuat, versi ke-1 tetap ada di history
- [ ] `PATCH /documents/:id/review` dengan decision `approved` → status `approved`
- [ ] `PATCH /documents/:id/review` oleh dosen bukan pembimbing → `403 Forbidden`
- [ ] `PATCH /documents/:id/review` saat status sudah `approved` → `422`
- [ ] `GET .../documents/:id/download` → return presigned URL (atau path lokal di dev)
- [ ] `GET .../documents/history?document_type=seminar_doc` → return semua versi
- [ ] `CanSubmitSeminar()` return `true` hanya jika `seminar_doc` berstatus `approved`
- [ ] `CanSubmitDefense()` return `true` hanya jika `defense_doc` berstatus `approved`
- [ ] Semua action tercatat di `audit_logs`
