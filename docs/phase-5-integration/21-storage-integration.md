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

- [ ] Install Supabase Go client:
  ```bash
  go get github.com/supabase-community/storage-go
  ```
- [ ] Buat `backend/pkg/storage/supabase_storage.go` — implementasi `StorageService` interface:
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

- [ ] Di `cmd/server/main.go`: inisiasi `SupabaseStorageService` dan inject ke handler dokumen
- [ ] Update `backend/internal/handler/document_handler.go`:
  - Upload file dari multipart request ke Supabase
  - Simpan `file_url` (path relatif) ke database
  - Bukan menyimpan URL publik langsung (agar akses dikontrol via presigned URL)
- [ ] Update `GET /documents/:id/download`:
  - Ambil `file_url` dari database
  - Generate presigned URL (expired 15 menit)
  - Return presigned URL ke client
- [ ] Update `GET /archives/:id/download`:
  - Generate presigned URL (expired 30 menit)

### Supabase Bucket Configuration

- [ ] Buat 2 bucket di Supabase dashboard:
  - `simtas-documents` — **private** (akses via presigned URL)
  - `simtas-archives` — **private** (akses via presigned URL)
- [ ] Set bucket policy: tidak ada akses publik langsung
- [ ] Catat bucket names di `.env`:
  ```env
  SUPABASE_DOCUMENTS_BUCKET=simtas-documents
  SUPABASE_ARCHIVES_BUCKET=simtas-archives
  ```

### Frontend — Upload dengan Progress

- [ ] Update `DocumentUploadModal.tsx` untuk upload dengan progress:
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

- [ ] Update semua tombol "Download" / "Lihat Dokumen":
  - Call `GET /documents/:id/download` → dapatkan `{ download_url, expires_in }`
  - Buka URL di tab baru: `window.open(downloadUrl, '_blank')`
  - Jangan simpan presigned URL di state terlalu lama (expired 15 menit)
- [ ] Loading state saat generate presigned URL

### Fallback — Development Mode

- [ ] Jika `STORAGE_PROVIDER=local` → gunakan stub storage (simpan ke `./tmp/uploads/`)
- [ ] Untuk development tanpa Supabase account
- [ ] Return URL lokal: `http://localhost:8080/tmp/uploads/{path}`
- [ ] Tambah static file handler di Gin untuk serve local uploads

### File Size Limit di Nginx (untuk deployment)

- [ ] Catat konfigurasi Nginx yang diperlukan (akan diterapkan di Job 24):
  ```nginx
  client_max_body_size 25M;  # untuk upload skripsi final
  ```

---

## Done Criteria

- [ ] Upload dokumen PDF → file tersimpan di Supabase bucket `simtas-documents`
- [ ] `GET /documents/:id/download` → presigned URL valid, file bisa dibuka di browser
- [ ] Presigned URL expired setelah 15 menit (test manual)
- [ ] Upload di frontend → progress bar berjalan 0–100%
- [ ] Cancel upload saat progress 50% → upload terhenti, tidak ada file di storage
- [ ] Upload arsip skripsi → file tersimpan di bucket `simtas-archives`
- [ ] `GET /archives/:id/download` → presigned URL valid 30 menit
- [ ] Mahasiswa lain coba download arsip yang bukan miliknya → `403 Forbidden`
- [ ] `STORAGE_PROVIDER=local` → upload ke folder lokal, download berfungsi di dev
- [ ] File yang dihapus (soft delete dokumen) tidak lagi bisa didownload
