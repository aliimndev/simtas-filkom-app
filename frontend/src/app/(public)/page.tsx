'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { useEffect, useState } from 'react'
import {
  ArrowRight,
  FileText,
  MessagesSquare,
  CalendarCheck,
  Archive,
  LayoutDashboard,
  Bell,
  ShieldCheck,
  Sparkles,
  ChevronDown,
} from 'lucide-react'
import { useAuthStore } from '@/lib/stores/auth-store'
import { Card, CardContent } from '@/components/ui/card'

const KEUNGGULAN = [
  {
    icon: LayoutDashboard,
    title: 'Terpusat & Transparan',
    desc: 'Seluruh proses TA & skripsi terdokumentasi di satu tempat yang dapat dipantau.',
  },
  {
    icon: ShieldCheck,
    title: 'Aman & Terkontrol',
    desc: 'Hak akses berbasis peran dengan pencatatan aktivitas (audit log) untuk keamanan.',
  },
  {
    icon: Sparkles,
    title: 'Alur Lengkap',
    desc: 'Dari pengajuan judul, bimbingan, seminar, hingga sidang—semua terintegrasi.',
  },
]

const FITUR = [
  {
    icon: FileText,
    title: 'Pengajuan Judul',
    desc: 'Mahasiswa mengajukan judul dan mendapat review dari kaprodi secara digital.',
  },
  {
    icon: MessagesSquare,
    title: 'Bimbingan Online',
    desc: 'Log bimbingan dengan dosen pembimbing tercatat dan dapat divalidasi.',
  },
  {
    icon: CalendarCheck,
    title: 'Jadwal Seminar & Sidang',
    desc: 'Penjadwalan, penguji, dan penilaian seminar serta sidang dalam satu alur.',
  },
  {
    icon: Archive,
    title: 'Arsip Digital',
    desc: 'Dokumen dan skripsi terdokumentasi serta dapat diakses kembali dengan mudah.',
  },
  {
    icon: LayoutDashboard,
    title: 'Dashboard Real-time',
    desc: 'Statistik dan status TA ditampilkan per peran: mahasiswa, dosen, kaprodi, admin.',
  },
  {
    icon: Bell,
    title: 'Notifikasi',
    desc: 'Pemberitahuan penting terkirim ke email pemangku kepentingan terkait.',
  },
]

const ALUR = [
  { step: '01', title: 'Pengajuan Judul', desc: 'Mahasiswa mengajukan judul skripsi.' },
  { step: '02', title: 'Review Judul', desc: 'Kaprodi menyetujui atau menolak dengan catatan.' },
  { step: '03', title: 'Penugasan Pembimbing', desc: 'Dosen pembimbing ditetapkan untuk mendampingi.' },
  { step: '04', title: 'Bimbingan & Dokumen', desc: 'Bimbingan rutin dan unggah proposal/skripsi.' },
  { step: '05', title: 'Seminar & Sidang', desc: 'Pelaksanaan seminar proposal hingga sidang akhir.' },
  { step: '06', title: 'Arsip & Kelulusan', desc: 'Skripsi diarsipkan dan status kelulusan dicatat.' },
]

const FAQ = [
  {
    q: 'Bagaimana cara login ke sistem?',
    a: 'Buka halaman Masuk, masukkan alamat email dan password yang telah diberikan oleh administrator fakultas.',
  },
  {
    q: 'Bagaimana jika saya lupa password?',
    a: 'Gunakan menu Lupa Password, masukkan email Anda, lalu ikuti tautan reset yang dikirim melalui email.',
  },
  {
    q: 'Bagaimana cara mengajukan judul skripsi?',
    a: 'Masuk sebagai mahasiswa, buka menu Skripsi Saya, lalu isi formulir pengajuan judul dan kirimkan untuk direview.',
  },
  {
    q: 'Apakah saya bisa melihat jadwal seminar dan sidang?',
    a: 'Ya. Jadwal seminar dan sidang tampil pada dashboard dan menu Jadwal sesuai peran masing-masing.',
  },
]

export default function HomePage() {
  const router = useRouter()
  const { accessToken, isHydrated } = useAuthStore()
  const [openFaq, setOpenFaq] = useState<number | null>(0)

  useEffect(() => {
    if (isHydrated && accessToken) {
      router.replace('/dashboard')
    }
  }, [isHydrated, accessToken, router])

  return (
    <div>
      {/* ── Hero ─────────────────────────────────────────────── */}
      <section className="landing-grid-bg border-b border-border">
        <div className="landing-container flex flex-col items-center py-24 text-center md:py-32">
          <span className="landing-eyebrow">
            <Sparkles className="h-4 w-4" /> Sistem Manajemen Tugas Akhir & Skripsi
          </span>
          <h1 className="landing-display mt-6 max-w-3xl text-4xl sm:text-5xl md:text-6xl">
            Kelola Tugas Akhir Anda dari Satu <span className="text-primary">Tempat</span>
          </h1>
          <p className="mt-6 max-w-2xl text-lg leading-relaxed text-muted-foreground">
            Platform resmi Fakultas Ilmu Komputer Universitas Djuanda untuk mengelola seluruh proses
            tugas akhir &amp; skripsi—pengajuan judul, bimbingan, seminar, sidang, hingga arsip.
          </p>
          <div className="mt-10 flex flex-col items-center gap-3 sm:flex-row">
            <Link
              href="/login"
              className="inline-flex h-12 items-center justify-center gap-2 rounded-lg bg-primary px-7 text-base font-medium text-primary-foreground shadow-sm transition hover:bg-primary-700"
            >
              Masuk ke Sistem <ArrowRight className="h-5 w-5" />
            </Link>
            <Link
              href="/guide"
              className="inline-flex h-12 items-center justify-center gap-2 rounded-lg border border-border px-7 text-base font-medium text-foreground transition hover:bg-muted"
            >
              Lihat Panduan
            </Link>
          </div>

          {/* Mock dashboard preview */}
          <div className="mt-16 w-full max-w-4xl">
            <div className="landing-surface rounded-2xl shadow-xl">
              <div className="flex items-center gap-2 border-b border-border px-5 py-3">
                <span className="h-3 w-3 rounded-full bg-danger/70" />
                <span className="h-3 w-3 rounded-full bg-warning/70" />
                <span className="h-3 w-3 rounded-full bg-success/70" />
                <span className="ml-3 text-xs text-muted-foreground">simtas.filkom.unida.ac.id/dashboard</span>
              </div>
              <div className="p-6">
                <div className="flex items-center justify-between">
                  <div>
                    <p className="text-xs font-medium text-muted-foreground">Selamat datang,</p>
                    <p className="text-lg font-bold text-foreground">Mahasiswa FILKOM</p>
                  </div>
                  <span className="rounded-full bg-success-50 px-3 py-1 text-xs font-medium text-success">
                    Dalam proses bimbingan
                  </span>
                </div>
                <div className="mt-6 grid gap-4 sm:grid-cols-3">
                  {[
                    { label: 'Judul', value: 'Disetujui' },
                    { label: 'Dokumen', value: '12' },
                    { label: 'Bimbingan', value: '8' },
                  ].map((c) => (
                    <div key={c.label} className="rounded-xl border border-border bg-background p-4 text-left">
                      <p className="text-xs text-muted-foreground">{c.label}</p>
                      <p className="mt-1 text-xl font-bold text-primary">{c.value}</p>
                    </div>
                  ))}
                </div>
              </div>
            </div>
          </div>
        </div>
      </section>


      {/* ── Keunggulan ───────────────────────────────────────── */}
      <section className="border-b border-border">
        <div className="landing-container py-20">
          <div className="mx-auto max-w-2xl text-center">
            <span className="landing-eyebrow justify-center">Keunggulan Sistem</span>
            <h2 className="landing-display mt-4 text-3xl md:text-4xl">Satu platform untuk seluruh proses akademik TA</h2>
          </div>
          <div className="mt-14 grid gap-6 md:grid-cols-3">
            {KEUNGGULAN.map((k) => (
              <Card key={k.title} className="transition hover:shadow-md">
                <CardContent className="p-6">
                  <span className="inline-flex h-11 w-11 items-center justify-center rounded-lg bg-primary-50 text-primary">
                    <k.icon className="h-5 w-5" />
                  </span>
                  <h3 className="mt-5 text-lg font-semibold text-foreground">{k.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{k.desc}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* ── Fitur ─────────────────────────────────────────────── */}
      <section className="landing-surface border-b border-border">
        <div className="landing-container py-20">
          <div className="mx-auto max-w-2xl text-center">
            <span className="landing-eyebrow justify-center">Fitur</span>
            <h2 className="landing-display mt-4 text-3xl md:text-4xl">Semua kebutuhan Tugas Akhir dalam satu alur</h2>
            <p className="mt-4 text-muted-foreground">
              Dirancang untuk kenyamanan mahasiswa, dosen, dan fakultas.
            </p>
          </div>
          <div className="mt-14 grid gap-6 sm:grid-cols-2 lg:grid-cols-3">
            {FITUR.map((f) => (
              <Card key={f.title} className="transition hover:shadow-md">
                <CardContent className="p-6">
                  <span className="inline-flex h-11 w-11 items-center justify-center rounded-lg bg-primary-50 text-primary">
                    <f.icon className="h-5 w-5" />
                  </span>
                  <h3 className="mt-5 text-lg font-semibold text-foreground">{f.title}</h3>
                  <p className="mt-2 text-sm leading-relaxed text-muted-foreground">{f.desc}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>


      {/* ── Alur Tugas Akhir ─────────────────────────────────── */}
      <section className="border-b border-border">
        <div className="landing-container py-20">
          <div className="mx-auto max-w-2xl text-center">
            <span className="landing-eyebrow justify-center">Alur Tugas Akhir</span>
            <h2 className="landing-display mt-4 text-3xl md:text-4xl">Lima tahap menuju kelulusan</h2>
          </div>
          <div className="mt-14 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {ALUR.map((a) => (
              <div key={a.step} className="flex gap-4 rounded-xl border border-border bg-background p-5">
                <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary text-sm font-bold text-primary-foreground">
                  {a.step}
                </span>
                <div>
                  <h3 className="font-semibold text-foreground">{a.title}</h3>
                  <p className="mt-1 text-sm text-muted-foreground">{a.desc}</p>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── FAQ Singkat ───────────────────────────────────────── */}
      <section className="landing-surface border-b border-border">
        <div className="landing-container py-20">
          <div className="mx-auto max-w-3xl">
            <div className="text-center">
              <span className="landing-eyebrow justify-center">FAQ</span>
              <h2 className="landing-display mt-4 text-3xl md:text-4xl">Pertanyaan yang sering diajukan</h2>
            </div>
            <div className="mt-10 space-y-3">
              {FAQ.map((item, i) => (
                <div key={item.q} className="overflow-hidden rounded-xl border border-border bg-background">
                  <button
                    type="button"
                    onClick={() => setOpenFaq(openFaq === i ? null : i)}
                    className="flex w-full items-center justify-between gap-4 px-5 py-4 text-left"
                  >
                    <span className="font-medium text-foreground">{item.q}</span>
                    <ChevronDown
                      className={`h-5 w-5 shrink-0 text-muted-foreground transition-transform ${openFaq === i ? 'rotate-180' : ''}`}
                    />
                  </button>
                  {openFaq === i && (
                    <div className="border-t border-border px-5 py-4 text-sm leading-relaxed text-muted-foreground">
                      {item.a}
                    </div>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* ── CTA Login ─────────────────────────────────────────── */}
      <section className="border-b border-border">
        <div className="landing-container py-24 text-center">
          <h2 className="landing-display mx-auto max-w-2xl text-3xl md:text-4xl">
            Siap memulai perjalanan Tugas Akhir Anda?
          </h2>
          <p className="mx-auto mt-4 max-w-xl text-muted-foreground">
            Masuk ke sistem untuk mengajukan judul, melanjutkan bimbingan, atau memantau proses.
          </p>
          <div className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link
              href="/login"
              className="inline-flex h-12 items-center justify-center gap-2 rounded-lg bg-primary px-7 text-base font-medium text-primary-foreground shadow-sm transition hover:bg-primary-700"
            >
              Masuk ke Sistem <ArrowRight className="h-5 w-5" />
            </Link>
            <Link
              href="/contact"
              className="inline-flex h-12 items-center justify-center gap-2 rounded-lg border border-border px-7 text-base font-medium text-foreground transition hover:bg-muted"
            >
              Hubungi Kami
            </Link>
          </div>
        </div>
      </section>
    </div>
  )
}

