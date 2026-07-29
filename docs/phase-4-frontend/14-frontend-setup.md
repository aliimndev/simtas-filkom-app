# Job 14 — Frontend Setup & Foundation

**Phase:** 4 — Frontend
**Referensi PRD:** Section 10 (Technology Stack), Section 20 (UI/UX Guidelines)
**Prerequisites:** Job 13 (Audit Log — Backend selesai penuh) ✅
**Estimasi:** 3 hari

---

## Objective

Menyiapkan fondasi frontend Next.js 16: design system, komponen UI reusable, layout per role, routing + protected routes, API client layer, dan auth state management. Setelah job ini selesai, semua halaman frontend bisa dibangun di atasnya tanpa setup ulang.

---

## Checklist

### Komponen UI Dasar (`src/components/ui/`)

Bangun semua komponen berikut dari scratch menggunakan Tailwind CSS (tidak pakai shadcn atau library UI eksternal — agar full kontrol atas desain):

- [ ] **Button** — variant: `primary`, `secondary`, `danger`, `ghost`; size: `sm`, `md`, `lg`; state: `loading`, `disabled`
- [ ] **Input** — text input dengan label, helper text, error state, required indicator
- [ ] **Textarea** — sama seperti Input tapi multiline
- [ ] **Select** — dropdown dengan label dan error state
- [ ] **Checkbox** dan **Radio**
- [ ] **Badge** — variant: `success`, `warning`, `danger`, `info`, `neutral` — untuk status
- [ ] **Alert** — inline alert dengan icon, variant success/warning/danger/info
- [ ] **Modal** — dialog dengan backdrop, close button, portal ke body
- [ ] **Table** — responsive table dengan header sticky, loading skeleton, empty state
- [ ] **Pagination** — navigasi halaman dengan prev/next dan page numbers
- [ ] **Spinner / LoadingOverlay**
- [ ] **Avatar** — inisial atau foto profil dengan fallback
- [ ] **Breadcrumb** — navigasi hierarki
- [ ] **Tooltip**
- [ ] **FileUpload** — drag-and-drop + click, preview nama file, validasi tipe dan ukuran
- [ ] **StatusBadge** — wrapper Badge khusus untuk status thesis:
  ```tsx
  const statusConfig: Record<string, { label: string; variant: BadgeVariant }> = {
    submitted:     { label: "Menunggu Review",   variant: "warning" },
    approved:      { label: "Disetujui",         variant: "success" },
    rejected:      { label: "Ditolak",           variant: "danger" },
    in_progress:   { label: "Bimbingan",         variant: "info" },
    seminar_ready: { label: "Siap Seminar",      variant: "info" },
    seminar_done:  { label: "Pasca Seminar",     variant: "info" },
    defense_ready: { label: "Siap Sidang",       variant: "info" },
    defense_done:  { label: "Pasca Sidang",      variant: "info" },
    graduated:     { label: "Lulus",             variant: "success" },
    cancelled:     { label: "Dibatalkan",        variant: "neutral" },
  }
  ```
- [ ] **Toast / Notification** — toast stack di sudut kanan bawah, auto-dismiss 4 detik, variant: success/error/warning/info
  - Implementasi menggunakan React context + portal
  - API: `useToast()` hook → `toast.success("...")`, `toast.error("...")`

### Design System (`src/lib/utils/`)

- [ ] **`cn(...classes)`** — merge Tailwind classes dengan `clsx` + `tailwind-merge`
- [ ] **`formatDate(date, format?)`** — format tanggal ke Bahasa Indonesia (misal: "15 Oktober 2026")
- [ ] **`formatDateTime(date)`** — format tanggal + waktu
- [ ] **`formatFileSize(bytes)`** — "1.2 MB", "450 KB"
- [ ] **`truncateText(text, maxLength)`**

### Tailwind Custom Theme (`tailwind.config.ts`)

- [ ] Definisikan warna brand FILKOM Unida:
  ```ts
  colors: {
    primary: {
      50: '#eff6ff',
      // ... (biru sebagai warna utama institusi)
      500: '#3b82f6',
      600: '#2563eb',
      700: '#1d4ed8',
    },
    // Tambahkan warna sekunder, success, warning, danger
  }
  ```
- [ ] Custom font: `Inter` dari Google Fonts (atau `Geist` dari Next.js default)
- [ ] Custom border radius, shadow, dan spacing jika diperlukan

### API Client Layer (`src/lib/api/`)

- [ ] **`src/lib/api/client.ts`** — Axios instance:
  ```ts
  const apiClient = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL,
    timeout: 30000,
  })

  // Request interceptor: attach access token
  apiClient.interceptors.request.use((config) => {
    const token = getAccessToken() // dari localStorage atau cookie
    if (token) config.headers.Authorization = `Bearer ${token}`
    return config
  })

  // Response interceptor: handle 401 (refresh token) dan error
  apiClient.interceptors.response.use(
    (res) => res,
    async (error) => {
      if (error.response?.status === 401) {
        // Coba refresh token
        // Jika gagal: redirect ke /login, clear tokens
      }
      return Promise.reject(error)
    }
  )
  ```

- [ ] **`src/lib/api/endpoints/`** — fungsi per modul:
  ```
  auth.ts        → login, logout, refreshToken, getMe, forgotPassword, resetPassword
  users.ts       → getUsers, createUser, updateUser, deleteUser, importUsers, getTemplate
  theses.ts      → getTheses, getThesis, createThesis, reviewThesis, assignSupervisor
  consultations.ts
  documents.ts
  seminars.ts
  defenses.ts
  archives.ts
  dashboard.ts
  academicYears.ts
  auditLogs.ts
  ```

- [ ] Setiap fungsi menggunakan `apiClient` dan return typed response:
  ```ts
  // Contoh: src/lib/api/endpoints/auth.ts
  export async function login(email: string, password: string): Promise<LoginResponse> {
    const res = await apiClient.post<ApiResponse<LoginResponse>>('/auth/login', { email, password })
    return res.data.data
  }
  ```

### TypeScript Types (`src/types/`)

- [ ] **`index.ts`** — re-export semua types
- [ ] **`auth.ts`** — `User`, `LoginRequest`, `LoginResponse`, `AuthState`
- [ ] **`thesis.ts`** — `Thesis`, `ThesisStatus`, `ThesisSupervisor`, `ThesisFilter`
- [ ] **`document.ts`** — `Document`, `DocumentType`, `DocumentStatus`
- [ ] **`consultation.ts`** — `ConsultationLog`, `ConsultationSummary`
- [ ] **`seminar.ts`** — `Seminar`, `SeminarScore`, `GradingComponent`
- [ ] **`defense.ts`** — `ThesisDefense`, `DefenseScore`
- [ ] **`archive.ts`** — `ThesisArchive`, `ArchiveFilter`
- [ ] **`dashboard.ts`** — semua dashboard response types
- [ ] **`api.ts`** — `ApiResponse<T>`, `PaginatedResponse<T>`, `PaginationMeta`
  ```ts
  interface ApiResponse<T> {
    success: boolean
    message?: string
    data: T
  }

  interface PaginatedResponse<T> {
    success: boolean
    data: T[]
    meta: PaginationMeta
  }

  interface PaginationMeta {
    page: number
    per_page: number
    total: number
    total_pages: number
  }
  ```

### Auth State Management (`src/lib/stores/`)

- [ ] Buat `src/lib/stores/auth-store.ts` menggunakan Zustand:
  ```ts
  interface AuthStore {
    user: User | null
    accessToken: string | null
    isAuthenticated: boolean
    isLoading: boolean
    setUser: (user: User) => void
    setTokens: (access: string, refresh: string) => void
    clearAuth: () => void
    initializeAuth: () => Promise<void>  // load dari localStorage saat app start
  }
  ```
- [ ] Install Zustand: `npm install zustand`
- [ ] Token disimpan di `localStorage` (access token) dan `httpOnly cookie` (refresh token — via backend Set-Cookie)
- [ ] `initializeAuth()` dipanggil di root layout saat mount pertama

### Layout & Routing (`src/app/`)

#### Root Layout (`src/app/layout.tsx`)
- [ ] Setup providers: QueryClientProvider, ToastProvider, AuthProvider
- [ ] Panggil `initializeAuth()` saat mount
- [ ] Font dan metadata aplikasi

#### Auth Layout (`src/app/(auth)/layout.tsx`)
- [ ] Layout minimal: centered card, logo FILKOM di atas
- [ ] Redirect ke `/dashboard` jika sudah login

#### Dashboard Layout (`src/app/(dashboard)/layout.tsx`)
- [ ] Redirect ke `/login` jika belum login
- [ ] Sidebar + topbar layout
- [ ] Sidebar menampilkan menu berbeda per role:
  ```ts
  const menuByRole: Record<UserRole, MenuItem[]> = {
    admin_fakultas: [
      { label: "Dashboard", href: "/dashboard", icon: HomeIcon },
      { label: "Manajemen User", href: "/admin/users", icon: UsersIcon },
      { label: "Tahun Akademik", href: "/admin/academic-years", icon: CalendarIcon },
      { label: "Jadwal", href: "/admin/schedules", icon: ClockIcon },
      { label: "Audit Log", href: "/admin/audit-logs", icon: ActivityIcon },
    ],
    kaprodi: [
      { label: "Dashboard", href: "/dashboard", icon: HomeIcon },
      { label: "Pengajuan Judul", href: "/kaprodi/thesis-reviews", icon: FileTextIcon },
      { label: "Seminar", href: "/kaprodi/seminars", icon: PresentationIcon },
      { label: "Sidang", href: "/kaprodi/defenses", icon: GraduationCapIcon },
      { label: "Jadwal", href: "/kaprodi/schedules", icon: CalendarIcon },
      { label: "Arsip", href: "/archives", icon: ArchiveIcon },
    ],
    mahasiswa: [
      { label: "Dashboard", href: "/dashboard", icon: HomeIcon },
      { label: "Skripsi Saya", href: "/student/thesis", icon: BookIcon },
      { label: "Bimbingan", href: "/student/consultations", icon: MessageCircleIcon },
      { label: "Dokumen", href: "/student/documents", icon: FolderIcon },
      { label: "Seminar", href: "/student/seminar", icon: PresentationIcon },
      { label: "Sidang", href: "/student/defense", icon: GraduationCapIcon },
      { label: "Arsip", href: "/archives", icon: ArchiveIcon },
    ],
    dosen_pembimbing: [
      { label: "Dashboard", href: "/dashboard", icon: HomeIcon },
      { label: "Mahasiswa Bimbingan", href: "/supervisor/students", icon: UsersIcon },
      { label: "Dokumen Review", href: "/supervisor/documents", icon: FolderIcon },
      { label: "Jadwal", href: "/supervisor/schedules", icon: CalendarIcon },
    ],
    dosen_penguji: [
      { label: "Dashboard", href: "/dashboard", icon: HomeIcon },
      { label: "Jadwal Pengujian", href: "/examiner/schedules", icon: CalendarIcon },
      { label: "Input Nilai", href: "/examiner/scoring", icon: StarIcon },
    ],
  }
  ```
- [ ] Topbar: nama user, role badge, avatar, dropdown (profil, ubah password, logout)
- [ ] Mobile: sidebar collapse menjadi bottom navigation atau hamburger menu
- [ ] Active menu item di-highlight sesuai current route

#### Protected Route Helper
- [ ] Buat `src/lib/hooks/useAuth.ts`:
  ```ts
  export function useRequireAuth(allowedRoles?: UserRole[]) {
    // Jika tidak authenticated → redirect /login
    // Jika role tidak diizinkan → redirect /dashboard dengan toast error
  }
  ```

### Custom Hooks (`src/lib/hooks/`)

- [ ] **`useToast()`** — akses toast context
- [ ] **`useAuth()`** — akses auth store
- [ ] **`usePagination(initialPage?, initialPerPage?)`** — state pagination
- [ ] **`useDebounce(value, delay)`** — debounce untuk search input
- [ ] **`useConfirm()`** — modal konfirmasi dengan Promise API:
  ```ts
  const { confirm } = useConfirm()
  const ok = await confirm({
    title: "Hapus User",
    message: "Apakah Anda yakin ingin menghapus user ini?",
    confirmLabel: "Hapus",
    variant: "danger"
  })
  if (ok) { /* lakukan hapus */ }
  ```

### Constants (`src/constants/`)

- [ ] **`routes.ts`** — semua path route
- [ ] **`queryKeys.ts`** — TanStack Query keys (untuk cache invalidation)
  ```ts
  export const queryKeys = {
    theses: { all: ['theses'], list: (f: ThesisFilter) => ['theses', 'list', f] },
    documents: { byThesis: (id: string) => ['documents', id] },
    // ...
  }
  ```

---

## Done Criteria

- [ ] `npm run dev` berjalan tanpa error atau TypeScript warning
- [ ] `npm run build` sukses (zero TS errors)
- [ ] Semua komponen UI ter-render di halaman test `/ui-preview` (hanya di development)
- [ ] Login → redirect ke `/dashboard` sesuai role
- [ ] Akses `/dashboard` tanpa login → redirect ke `/login`
- [ ] Sidebar menampilkan menu yang berbeda untuk setiap role
- [ ] Toast muncul dan auto-dismiss setelah 4 detik
- [ ] API client mengirim `Authorization: Bearer <token>` di setiap request
- [ ] Expired token → auto-refresh, jika refresh gagal → redirect login
- [ ] Responsive: sidebar collapse di mobile (layar < 768px)
