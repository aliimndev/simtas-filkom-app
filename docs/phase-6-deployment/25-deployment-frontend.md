# Job 25 — Deployment Frontend (Vercel)

**Phase:** 6 — Deployment
**Referensi PRD:** Section 11 (Deployment — Frontend: Vercel)
**Prerequisites:** Job 24 (Deployment Backend) ✅
**Estimasi:** 1 hari

---

## Objective

Deploy frontend Next.js ke Vercel, konfigurasi environment variables production, dan pastikan domain production terhubung dengan benar ke backend API. Setelah job ini selesai, aplikasi dapat diakses publik di domain resmi FILKOM.

---

## Checklist

### Setup Vercel Project

- [ ] Login ke Vercel dashboard → New Project → Import dari GitHub repo
- [ ] Pilih folder `frontend/` sebagai root directory
- [ ] Framework: Next.js (Vercel auto-detect)
- [ ] Build command: `npm run build`
- [ ] Output directory: `.next`

### Environment Variables di Vercel

- [ ] Set di Vercel dashboard → Settings → Environment Variables:
  ```
  NEXT_PUBLIC_API_URL = https://api.simtas.filkom.unida.ac.id/api/v1
  NEXT_PUBLIC_APP_NAME = SIMTAS FILKOM
  NEXT_PUBLIC_APP_VERSION = 1.0.0
  ```
- [ ] Pastikan prefix `NEXT_PUBLIC_` untuk semua variable yang diakses di browser
- [ ] Variable tanpa prefix hanya tersedia di server-side (tidak ada di v1.0)

### Custom Domain

- [ ] Di Vercel dashboard → Settings → Domains:
  - Tambah domain: `simtas.filkom.unida.ac.id`
- [ ] Di DNS provider (cPanel/Cloudflare FILKOM):
  - Tambah CNAME record: `simtas` → `cname.vercel-dns.com`
  - Atau A record jika menggunakan IP
- [ ] Vercel otomatis provision SSL via Let's Encrypt
- [ ] Pastikan domain `api.simtas.filkom.unida.ac.id` (backend) sudah aktif dari Job 24

### CORS Verification

- [ ] Pastikan backend `.env.production` sudah set:
  ```env
  CORS_ALLOWED_ORIGINS=https://simtas.filkom.unida.ac.id
  ```
- [ ] Test dari browser: login di `https://simtas.filkom.unida.ac.id` → tidak ada CORS error di console

### next.config.ts Production Settings

- [ ] Review dan pastikan konfigurasi ini ada:
  ```ts
  const nextConfig: NextConfig = {
    images: {
      domains: ['xxx.supabase.co'],
      formats: ['image/avif', 'image/webp'],
    },
    // Security headers
    async headers() {
      return [
        {
          source: '/(.*)',
          headers: [
            { key: 'X-Frame-Options', value: 'DENY' },
            { key: 'X-Content-Type-Options', value: 'nosniff' },
            { key: 'Referrer-Policy', value: 'strict-origin-when-cross-origin' },
          ],
        },
      ]
    },
    // Redirect / ke /dashboard atau /login
    async redirects() {
      return [
        { source: '/', destination: '/dashboard', permanent: false },
      ]
    },
  }
  ```

### Vercel Deployment Settings

- [ ] Branch deployment:
  - `main` → production (`simtas.filkom.unida.ac.id`)
  - `develop` → preview (`simtas-git-develop-xxx.vercel.app`) — untuk staging
- [ ] Build & deployment notification ke Slack/email (opsional)
- [ ] Enable "Automatically expose System Environment Variables"

### Post-Deploy Checklist

- [ ] Buka `https://simtas.filkom.unida.ac.id` di browser → halaman login muncul
- [ ] Login dengan akun admin → dashboard muncul
- [ ] Cek Network tab di DevTools: request ke `api.simtas.filkom.unida.ac.id` status 200
- [ ] Tidak ada mixed content warning (semua HTTP resource sudah HTTPS)
- [ ] Lighthouse score: Performance ≥ 80, Accessibility ≥ 90

---

## Done Criteria

- [ ] `https://simtas.filkom.unida.ac.id` terbuka dengan halaman login
- [ ] Login berfungsi (terhubung ke backend production)
- [ ] Tidak ada CORS error di browser console
- [ ] SSL certificate valid untuk domain frontend
- [ ] Push ke branch `main` → Vercel otomatis deploy (auto-deploy aktif)
- [ ] Environment variables production sudah di-set di Vercel (bukan hardcoded di kode)
- [ ] Aplikasi responsive di mobile (test di Chrome DevTools)
