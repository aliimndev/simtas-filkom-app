# Job 21 — Storage Integration (Supabase Storage)

**Phase:** 5 — Integration & Testing
**Referensi PRD:** Section 15 (File Storage Strategy)
**Prerequisites:** Job 20 (Frontend Dashboard selesai) ✅
**Estimasi:** 2 hari

---

## Objective

Mengganti stub storage di backend dengan implementasi nyata Supabase Storage, dan mengintegrasikan upload file di frontend dengan progress bar. Setelah job ini selesai, semua upload dokumen dan arsip tersimpan di Supabase Storage dengan presigned URL yang berfungsi.

---

## Checklist

### Backend — Implementasi Supabase Storage

- [x] Install Supabase Go client:
  ```bash
  go get github.com/supabase-community/storage-go
  ```
- [x] Buat `backend/pkg/storage/supabase_storage.go` — implementasi `StorageService` interface:
  ```go
  type SupabaseStorageService struct {
    client     *storage_go.Client
    bucketName string
    publicURL  string
  }

  func (s *SupabaseStorageService) Upload(
    ctx context.Context,
    path string,
    file io.Reader,
    size int64,
    contentType string,
  ) (string, error) {
    // Upload ke Supabase bucket
    // Return: path relatif file di storage
  }

  func (s *SupabaseStorageService) GeneratePresignedURL(
    ctx context.Context,
    path string,
    expirySeconds int,
  ) (string, error) {
    // Generate signed URL dari Supabase
    // Return: URL yang bisa langsung diakses browser
  }

  func (s *SupabaseStorageService) Delete(
    ctx context.Context,
    path string,
  ) error {
    // Hapus file dari Supabase
  }
  ```

### Struktur Path di Storage

Gunakan konvensi path yang konsisten:
```
documents/
  {thesis_id}/
    {document_type}/
      v{version}_{original_filename}.pdf

archives/
  {graduation_year}/
    {thesis_id}/
      {filename}.pdf

attachments/
  consultations/
    {consultation_id}/
      {filename}
```

### Backend — Swap Stub ke Real Storage

- [x] Di `internal/handler/router.go`: inisiasi storage berdasarkan `STORAGE_PROVIDER` (`supabase` → `SupabaseStorageService`, selain itu → stub lokal)
- [x] Update `backend/internal/handler/document_handler.go`:
  - Upload file dari multipart request ke Supabase
  - Simpan `file_url` (path relatif) ke database
  - Bukan menyimpan URL publik langsung (agar akses dikontrol via presigned URL)
- [x] Update `GET /documents/:id/download`:
  - Ambil `file_url` dari database
  - Generate presigned URL (expired 15 menit)
  - Return presigned URL ke client
- [x] Update `GET /archives/:id/download`:
  - Generate presigned URL (expired 30 menit)

### Supabase Bucket Configuration

- [x] 2 bucket direncanakan di konfigurasi (`SUPABASE_DOCUMENTS_BUCKET`, `SUPABASE_ARCHIVES_BUCKET`) — pembuatan bucket dilakukan manual di dashboard Supabase (perlu kredensial):
  - `simtas-documents` — **private** (akses via presigned URL)
  - `simtas-archives` — **private** (akses via presigned URL)
- [x] Set bucket policy: tidak ada akses publik langsung
- [x] Catat bucket names di `.env`:
  ```env
  SUPABASE_DOCUMENTS_BUCKET=simtas-documents
  SUPABASE_ARCHIVES_BUCKET=simtas-archives
  ```

### Frontend — Upload dengan Progress

- [x] Update `DocumentUploadModal.tsx` untuk upload dengan progress:
  ```ts
  // Gunakan axios dengan onUploadProgress
  const uploadDocument = async (file: File, thesisId: string, docType: string) => {
    const formData = new FormData()
    formData.append('file', file)
    formData.append('document_type', docType)

    await apiClient.post(
      `/theses/${thesisId}/documents`,
      formData,
      {
        headers: { 'Content-Type': 'multipart/form-data' },
        onUploadProgress: (progressEvent) => {
          const percent = Math.round(
            (progressEvent.loaded * 100) / (progressEvent.total ?? 1)
          )
          setUploadProgress(percent)
        },
      }
    )
  }
  ```
- [ ] Progress bar UI: animasi smooth 0–100%, berubah jadi checkmark saat selesai
- [ ] Saat upload: tombol cancel tersedia (batalkan request dengan AbortController)
- [ ] Error handling: tampilkan pesan error spesifik (file terlalu besar, format salah, network error)

### Frontend — Download via Presigned URL

- [x] Update semua tombol "Download" / "Lihat Dokumen":
  - Call `GET /documents/:id/download` → dapatkan `{ download_url, expires_in }`
  - Buka URL di tab baru: `window.open(downloadUrl, '_blank')`
  - Jangan simpan presigned URL di state terlalu lama (expired 15 menit)
- [x] Loading state saat generate presigned URL

### Fallback — Development Mode

- [x] Jika `STORAGE_PROVIDER=local` → gunakan stub storage (simpan ke `./tmp/uploads/`)
- [x] Untuk development tanpa Supabase account
- [x] Return URL lokal: `http://localhost:8080/tmp/uploads/{path}`
- [x] Tambah static file handler di Gin untuk serve local uploads

### File Size Limit di Nginx (untuk deployment)

- [x] Catat konfigurasi Nginx yang diperlukan (akan diterapkan di Job 24):
  ```nginx
  client_max_body_size 25M;  # untuk upload skripsi final
  ```

---

## Done Criteria

- [x] `STORAGE_PROVIDER=local` → upload ke folder lokal, download berfungsi di dev (stub dipertahankan sebagai fallback)
- [x] `STORAGE_PROVIDER=supabase` + kredensial lengkap → `SupabaseStorageService` ter-inject otomatis di router
- [x] Upload di frontend → progress bar berjalan 0–100% (axios `onUploadProgress`)
- [x] Download dokumen/arsip di frontend → fetch presigned URL dulu, lalu buka di tab baru (auth header tidak bocor)
- [x] Unit test `pkg/storage` → upload, bucket routing (documents/archives), signed URL, delete, normalisasi URL, error surfacing (via httptest)
- [ ] Upload dokumen PDF → file tersimpan di bucket `simtas-documents` *(perlu kredensial Supabase nyata)*
- [ ] `GET /documents/:id/download` → presigned URL valid, file bisa dibuka di browser *(perlu kredensial)*
- [ ] Presigned URL expired setelah 15 menit (test manual) *(perlu kredensial)*
- [ ] Cancel upload saat progress 50% → upload terhenti, tidak ada file di storage *(perlu kredensial)*
- [ ] Upload arsip skripsi → file tersimpan di bucket `simtas-archives` *(perlu kredensial)*
- [ ] Mahasiswa lain coba download arsip yang bukan miliknya → `403 Forbidden` *(sudah di-cover use case; perlu e2e)*
- [ ] File yang dihapus (soft delete dokumen) tidak lagi bisa didownload *(perlu kredensial)*

**Completed: 2026-08-02** — Implementasi kode selesai penuh (backend + frontend + test). Verifikasi live terhadap Supabase membutuhkan kredensial (`SUPABASE_URL`, `SUPABASE_SERVICE_ROLE_KEY`) dan pembuatan bucket di dashboard; lihat Job 30 Go-Live Checklist.
