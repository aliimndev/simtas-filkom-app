# RESOURCES — Sumber Belajar Object Storage

Sumber resmi & berotoritas tinggi. Dipakai untuk men-grounded-kan semua materi
pembelajaran (jangan pernah hanya andalkan pengetahuan parametrik).

## Landasan konsep

- **AWS S3 — What is object storage?** (dasar konsep)
  https://aws.amazon.com/s3/
- **AWS S3 — User Guide (bucket, object, presigned URL, ACK/signature)**
  https://docs.aws.amazon.com/AmazonS3/latest/userguide/
- **Cloudflare — R2 S3 API compatibility** (standar S3 lintas vendor)
  https://developers.cloudflare.com/r2/api/s3/
- **MinIO — S3 API (dokumentasi API kompatibel S3)**
  https://min.io/docs/minio/linux/developers/go/API.html

## Perbandingan & harga (sumber resmi)

- **Cloudflare R2 pricing** (free tier, Class A/B ops, egress gratis)
  https://developers.cloudflare.com/r2/pricing/
- **Backblaze B2 pricing** (S3-compatible, $6.95/TB-bln, egress 3x storage gratis)
  https://www.backblaze.com/cloud-storage/pricing
- **AWS S3 pricing** (harga /GB-bulan, egress/transfer keluar)
  https://aws.amazon.com/s3/pricing/
- **MinIO Community Edition** (self-hosted gratis, AGPL)
  https://min.io/product/overview

## Integrasi dengan stack SIMTAS (Go)

- **minio-go** (Go SDK untuk MinIO, juga bisa ke S3-compatible lain)
  https://github.com/minio/minio-go
- **aws-sdk-go-v2** (SDK AWS resmi untuk Go; kompatibel dengan S3 & R2 via endpoint)
  https://github.com/aws/aws-sdk-go-v2
- **Supabase Storage (S3 compatible endpoint)** — untuk memahami "Supabase juga
  pada dasarnya object storage S3"
  https://supabase.com/docs/guides/storage

## Sumber dalam project (konteks nyata)

- `docs/phase-5-integration/21-storage-integration.md` — Job 21 (storage Supabase)
- `docs/phase-6-deployment/24-deployment-backend.md` — konfigurasi produksi
- `backend/internal/domain/service/storage_service.go` — interface `StorageService`
- `backend/pkg/storage/*.go` — implementasi stub & Supabase
- `backend/internal/handler/router.go` — pilihan `STORAGE_PROVIDER`
