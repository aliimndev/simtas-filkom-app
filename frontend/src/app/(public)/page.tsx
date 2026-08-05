'use client'

import Link from 'next/link'
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
  ChevronDown,
  Users,
  GraduationCap,
} from 'lucide-react'
import { useRouter } from 'next/navigation'
import { useAuthStore } from '@/lib/stores/auth-store'
import { Reveal } from '@/components/features/landing/reveal'

const ROLES = ['mahasiswa', 'dosen pembimbing', 'dosen penguji', 'kaprodi', 'admin fakultas']

const STATS = [
  { value: '5', label: 'Peran pengguna' },
  { value: '6', label: 'Tahap Tugas Akhir' },
  { value: '1', label: 'Fakultas Ilmu Komputer' },
]

const ALUR = [
  { step: '01', title: 'Pengajuan Judul', desc: 'Mahasiswa mengajukan judul Tugas Akhir Skripsi secara digital.' },
  { step: '02', title: 'Review Judul', desc: 'Kaprodi menyetujui atau menolak beserta catatan.' },
  { step: '03', title: 'Penugasan Pembimbing', desc: 'Dosen pembimbing ditetapkan untuk mendampingi.' },
  { step: '04', title: 'Bimbingan & Dokumen', desc: 'Log bimbingan tercatat dan dokumen diunggah bertahap.' },
  { step: '05', title: 'Seminar & Sidang', desc: 'Seminar proposal hingga sidang akhir terjadwal.' },
  { step: '06', title: 'Arsip & Kelulusan', desc: 'Dokumen final diarsipkan dan status kelulusan dicatat.' },
]

const FITUR = [
  { icon: FileText, title: 'Pengajuan Judul', desc: 'Ajukan judul dan dapatkan review dari kaprodi secara digital.' },
  { icon: MessagesSquare, title: 'Bimbingan Online', desc: 'Catat log bimbingan dan menunggu validasi dosen pembimbing.' },
  { icon: CalendarCheck, title: 'Seminar & Sidang', desc: 'Penjadwalan, penguji, dan penilaian dalam satu alur.' },
  { icon: Archive, title: 'Arsip Digital', desc: 'Dokumen final terdokumentasi dan mudah diakses kembali.' },
  { icon: LayoutDashboard, title: 'Dashboard Real-time', desc: 'Statistik dan status per peran: mahasiswa, dosen, kaprodi, admin.' },
  { icon: Bell, title: 'Notifikasi Email', desc: 'Pemberitahuan penting terkirim ke pemangku kepentingan.' },
]

const PENGGUNA = [
  { icon: GraduationCap, role: 'Mahasiswa', desc: 'Mengajukan judul, mengunggah dokumen, dan memantau bimbingan.' },
  { icon: Users, role: 'Dosen Pembimbing & Penguji', desc: 'Memvalidasi bimbingan, mereview dokumen, dan menilai.' },
  { icon: ShieldCheck, role: 'Kaprodi', desc: 'Menyetujui judul, memantau progres, dan mengelola penjadwalan.' },
  { icon: LayoutDashboard, role: 'Administrator Fakultas', desc: 'Mengelola pengguna, konfigurasi sistem, dan data master.' },
]

const FAQ = [
  { q: 'Bagaimana cara login ke sistem?', a: 'Buka halaman Masuk, lalu gunakan alamat email dan password yang diberikan oleh administrator fakultas.' },
  { q: 'Bagaimana jika saya lupa password?', a: 'Gunakan menu Lupa Password, masukkan email Anda, lalu ikuti tautan reset yang dikirim melalui email.' },
  { q: 'Bagaimana cara mengajukan judul?', a: 'Masuk sebagai mahasiswa, buka menu Tugas Akhir Skripsi, lalu isi formulir pengajuan judul dan kirimkan untuk direview.' },
  { q: 'Apakah saya bisa melihat jadwal seminar dan sidang?', a: 'Ya. Jadwal tampil pada dashboard dan menu Jadwal sesuai peran masing-masing.' },
]

function RoleCycler() {
  const [i, setI] = useState(0)
  useEffect(() => {
    const id = setInterval(() => setI((v) => (v + 1) % ROLES.length), 2200)
    return () => clearInterval(id)
  }, [])
  return (
    <span className="font-display italic text-[var(--st-text)]">
      <span key={i} className="animate-st-role inline-block">{ROLES[i]}</span>
    </span>
  )
}

export default function HomePage() {
  const router = useRouter()
  const { accessToken, isHydrated } = useAuthStore()
  const [openFaq, setOpenFaq] = useState<number | null>(0)

  useEffect(() => {
    if (isHydrated && accessToken) router.replace('/dashboard')
  }, [isHydrated, accessToken, router])

  return (
    <div>
      {/* ── Hero ─────────────────────────────────────────────── */}
      <section className="landing-grid-bg relative overflow-hidden border-b border-[var(--st-stroke)]">
        {/* accent glow */}
        <div
          aria-hidden
          className="pointer-events-none absolute left-1/2 top-[-10%] h-[40rem] w-[40rem] -translate-x-1/2 rounded-full opacity-20 blur-[120px]"
          style={{ background: 'radial-gradient(circle, var(--st-accent-to), transparent 60%)' }}
        />
        <div className="landing-container relative flex flex-col items-center pt-36 pb-24 text-center md:pt-44 md:pb-32">
          <Reveal>
            <span className="landing-eyebrow">
              SIMTAS://FILKOM <span className="text-[var(--st-stroke)]">·</span> Universitas Djuanda
            </span>
          </Reveal>
          <Reveal delay={80}>
            <h1 className="landing-display mt-7 max-w-4xl text-[2.6rem] leading-[0.98] sm:text-6xl md:text-7xl">
              Satu sistem untuk seluruh perjalanan <span className="accent-text italic">Tugas Akhir Skripsi</span>.
            </h1>
          </Reveal>
          <Reveal delay={160}>
            <p className="mt-7 flex items-center justify-center gap-2 text-base text-[var(--st-muted)] md:text-lg">
              Dibangun untuk <RoleCycler />.
            </p>
          </Reveal>
          <Reveal delay={240}>
            <p className="mt-6 max-w-xl text-sm leading-relaxed text-[var(--st-muted)] md:text-base">
              Dari pengajuan judul, bimbingan, seminar, hingga sidang dan arsip—seluruh proses
              Tugas Akhir Skripsi Fakultas Ilmu Komputer dalam satu ekosistem digital.
            </p>
          </Reveal>
          <Reveal delay={320}>
            <div className="mt-10 flex flex-col items-center gap-3 sm:flex-row">
              <Link
                href="/login"
                className="accent-ring inline-flex h-12 items-center justify-center gap-2 rounded-full bg-[var(--st-text)] px-7 text-sm font-medium text-[var(--st-bg)] transition hover:opacity-90"
              >
                Masuk ke Sistem <ArrowRight className="h-4 w-4" />
              </Link>
              <a
                href="#alur"
                className="accent-ring inline-flex h-12 items-center justify-center gap-2 rounded-full border border-[var(--st-stroke)] bg-[var(--st-surface)] px-7 text-sm font-medium text-[var(--st-text)] transition hover:border-[var(--st-accent-from)]/40 hover:bg-[var(--st-surface-hi)]"
              >
                Lihat Alur
              </a>
            </div>
          </Reveal>
        </div>

        {/* scroll indicator */}
        <div className="flex flex-col items-center gap-2 pb-8">
          <span className="font-mono text-[0.65rem] uppercase tracking-[0.25em] text-[var(--st-muted)]">scroll</span>
          <span className="relative h-10 w-px overflow-hidden bg-[var(--st-stroke)]">
            <span className="accent-gradient absolute inset-x-0 top-0 h-1/3 animate-st-scroll" />
          </span>
        </div>
      </section>

      {/* ── Stats strip ──────────────────────────────────────── */}
      <section className="border-b border-[var(--st-stroke)]">
        <div className="landing-container grid grid-cols-3 divide-x divide-[var(--st-stroke)]">
          {STATS.map((s) => (
            <div key={s.label} className="px-3 py-10 text-center md:py-12">
              <p className="font-display text-4xl text-[var(--st-text)] md:text-6xl">{s.value}</p>
              <p className="mt-2 text-[0.7rem] uppercase tracking-[0.2em] text-[var(--st-muted)] md:text-xs">{s.label}</p>
            </div>
          ))}
        </div>
      </section>

      {/* ── Pipeline (signature) ────────────────────────────── */}
      <section id="alur" className="border-b border-[var(--st-stroke)]">
        <div className="landing-container py-20 md:py-28">
          <Reveal>
            <span className="landing-eyebrow"><span className="h-px w-8 bg-[var(--st-stroke)]" /> Alur Tugas Akhir</span>
          </Reveal>
          <Reveal delay={80}>
            <h2 className="landing-display mt-5 max-w-2xl text-3xl md:text-5xl">
              Enam tahap, satu <span className="accent-text italic">jejak</span> utuh.
            </h2>
          </Reveal>
          <Reveal delay={160}>
            <p className="mt-4 max-w-xl text-sm leading-relaxed text-[var(--st-muted)] md:text-base">
              Setiap mahasiswa melalui urutan yang sama—urutan ini bukan dekorasi, melainkan proses
              nyata yang menentukan maju atau tertunda.
            </p>
          </Reveal>

          <div className="mt-14 grid gap-px overflow-hidden rounded-2xl border border-[var(--st-stroke)] bg-[var(--st-stroke)] sm:grid-cols-2 lg:grid-cols-3">
            {ALUR.map((a, i) => (
              <Reveal key={a.step} delay={i * 60} className="h-full">
                <div className="group relative flex h-full flex-col gap-3 bg-[var(--st-surface)] p-6 transition-colors hover:bg-[var(--st-surface-hi)] md:p-8">
                  <div className="flex items-baseline justify-between">
                    <span className="font-display text-4xl text-[var(--st-muted)] transition-colors group-hover:text-[var(--st-text)] md:text-5xl">{a.step}</span>
                    {i < ALUR.length - 1 && (
                      <ArrowRight className="h-4 w-4 text-[var(--st-stroke)] transition-colors group-hover:text-[var(--st-accent-from)]" />
                    )}
                  </div>
                  <h3 className="text-base font-medium text-[var(--st-text)] md:text-lg">{a.title}</h3>
                  <p className="text-sm leading-relaxed text-[var(--st-muted)]">{a.desc}</p>
                </div>
              </Reveal>
            ))}
          </div>
        </div>
      </section>

      {/* ── Fitur (bento) ────────────────────────────────────── */}
      <section className="landing-surface border-b border-[var(--st-stroke)]">
        <div className="landing-container py-20 md:py-28">
          <Reveal>
            <span className="landing-eyebrow"><span className="h-px w-8 bg-[var(--st-stroke)]" /> Kemampuan</span>
          </Reveal>
          <Reveal delay={80}>
            <h2 className="landing-display mt-5 text-3xl md:text-5xl">Yang bisa Anda <span className="accent-text italic">lakukan</span>.</h2>
          </Reveal>

          <div className="mt-14 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
            {FITUR.map((f, i) => {
              const Icon = f.icon
              return (
                <Reveal key={f.title} delay={i * 60} className="h-full">
                  <div className="group st-card flex h-full flex-col gap-4 rounded-2xl p-6 transition-colors hover:border-[var(--st-accent-from)]/40 md:p-7">
                    <span className="flex h-10 w-10 items-center justify-center rounded-xl border border-[var(--st-stroke)] bg-[var(--st-bg)] text-[var(--st-accent-from)] transition-colors group-hover:text-[var(--st-accent-to)]">
                      <Icon className="h-5 w-5" />
                    </span>
                    <h3 className="text-base font-medium text-[var(--st-text)]">{f.title}</h3>
                    <p className="text-sm leading-relaxed text-[var(--st-muted)]">{f.desc}</p>
                  </div>
                </Reveal>
              )
            })}
          </div>
        </div>
      </section>

      {/* ── Peran pengguna ───────────────────────────────────── */}
      <section className="border-b border-[var(--st-stroke)]">
        <div className="landing-container py-20 md:py-28">
          <Reveal>
            <span className="landing-eyebrow"><span className="h-px w-8 bg-[var(--st-stroke)]" /> Pengguna</span>
          </Reveal>
          <Reveal delay={80}>
            <h2 className="landing-display mt-5 max-w-2xl text-3xl md:text-5xl">Dibuat untuk setiap <span className="accent-text italic">peran</span>.</h2>
          </Reveal>

          <div className="mt-14 grid gap-4 sm:grid-cols-2">
            {PENGGUNA.map((p, i) => {
              const Icon = p.icon
              return (
                <Reveal key={p.role} delay={i * 60} className="h-full">
                  <div className="st-card flex h-full items-start gap-5 rounded-2xl p-6 md:p-7">
                    <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl accent-gradient text-[var(--st-bg)]">
                      <Icon className="h-5 w-5" />
                    </span>
                    <div>
                      <h3 className="text-base font-medium text-[var(--st-text)]">{p.role}</h3>
                      <p className="mt-1.5 text-sm leading-relaxed text-[var(--st-muted)]">{p.desc}</p>
                    </div>
                  </div>
                </Reveal>
              )
            })}
          </div>
        </div>
      </section>

      {/* ── FAQ ──────────────────────────────────────────────── */}
      <section className="landing-surface border-b border-[var(--st-stroke)]">
        <div className="landing-container py-20 md:py-28">
          <Reveal>
            <div className="max-w-2xl">
              <span className="landing-eyebrow">FAQ</span>
              <h2 className="landing-display mt-5 text-3xl md:text-5xl">Pertanyaan yang sering muncul.</h2>
            </div>
          </Reveal>
          <div className="mt-10 max-w-2xl">
            {FAQ.map((item, i) => (
              <div key={item.q} className="border-b border-[var(--st-stroke)]">
                <button
                  type="button"
                  onClick={() => setOpenFaq(openFaq === i ? null : i)}
                  aria-expanded={openFaq === i}
                  aria-controls={`faq-panel-${i}`}
                  className="flex w-full items-center justify-between gap-4 py-5 text-left"
                >
                  <span className="text-sm font-medium text-[var(--st-text)] md:text-base">{item.q}</span>
                  <ChevronDown className={`h-4 w-4 shrink-0 text-[var(--st-muted)] transition-transform ${openFaq === i ? 'rotate-180' : ''}`} />
                </button>
                {openFaq === i && (
                  <p id={`faq-panel-${i}`} className="pb-5 text-sm leading-relaxed text-[var(--st-muted)]">{item.a}</p>
                )}
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ── CTA + marquee ────────────────────────────────────── */}
      <section className="relative overflow-hidden">
        <div className="landing-container flex flex-col items-center py-24 text-center md:py-32">
          <Reveal>
            <h2 className="landing-display mx-auto max-w-2xl text-3xl md:text-5xl">
              Siap memulai <span className="accent-text italic">Tugas Akhir</span> Anda?
            </h2>
          </Reveal>
          <Reveal delay={80}>
            <p className="mx-auto mt-5 max-w-xl text-sm leading-relaxed text-[var(--st-muted)] md:text-base">
              Masuk untuk mengajukan judul, melanjutkan bimbingan, atau memantau proses.
            </p>
          </Reveal>
          <Reveal delay={160}>
            <div className="mt-9 flex flex-col items-center gap-3 sm:flex-row">
              <Link
                href="/login"
                className="accent-ring inline-flex h-12 items-center justify-center gap-2 rounded-full bg-[var(--st-text)] px-7 text-sm font-medium text-[var(--st-bg)] transition hover:opacity-90"
              >
                Masuk ke Sistem <ArrowRight className="h-4 w-4" />
              </Link>
              <Link
                href="/contact"
                className="accent-ring inline-flex h-12 items-center justify-center gap-2 rounded-full border border-[var(--st-stroke)] bg-[var(--st-surface)] px-7 text-sm font-medium text-[var(--st-text)] transition hover:border-[var(--st-accent-from)]/40 hover:bg-[var(--st-surface-hi)]"
              >
                Hubungi Kami
              </Link>
            </div>
          </Reveal>
        </div>

        {/* marquee */}
        <div className="border-t border-[var(--st-stroke)] py-5">
          <div className="flex w-max animate-st-marquee whitespace-nowrap">
            {Array.from({ length: 2 }).map((_, dup) => (
              <div key={dup} className="flex items-center gap-8 pr-8" aria-hidden={dup === 1}>
                {['PENGAJUAN', 'BIMBINGAN', 'SEMINAR', 'SIDANG', 'ARSIP'].map((w) => (
                  <span key={w} className="font-display text-2xl italic text-[var(--st-muted)] md:text-3xl">
                    {w} <span className="text-[var(--st-accent-from)]">•</span>
                  </span>
                ))}
              </div>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}
