import { Target, GraduationCap, Cpu, CheckCircle2, Users, LayoutDashboard, ShieldCheck } from 'lucide-react'
import { Reveal } from '@/components/features/landing/reveal'

const TUJUAN = [
  'Mempermudah administrasi Tugas Akhir Skripsi yang sebelumnya manual.',
  'Menjamin transparansi proses dari pengajuan judul hingga kelulusan.',
  'Mendigitalisasi arsip dan dokumen akademik agar mudah diakses kembali.',
]

const MANFAAT = [
  'Mahasiswa dapat memantau status Tugas Akhir secara real-time.',
  'Dosen pembimbing dan penguji mendapat alur validasi yang jelas.',
  'Kaprodi dan admin memperoleh laporan serta audit yang ringkas.',
  'Mengurangi dokumen fisik dan mempercepat alur persetujuan.',
]

const PENGGUNA = [
  { icon: GraduationCap, role: 'Mahasiswa', desc: 'Mengajukan judul, mengunggah dokumen, dan memantau bimbingan.' },
  { icon: Users, role: 'Dosen Pembimbing & Penguji', desc: 'Memvalidasi bimbingan, mereview dokumen, dan melakukan penilaian.' },
  { icon: ShieldCheck, role: 'Kaprodi', desc: 'Menyetujui judul, memantau progres, dan mengelola penjadwalan.' },
  { icon: LayoutDashboard, role: 'Administrator Fakultas', desc: 'Mengelola pengguna, konfigurasi sistem, dan data master.' },
]

const TEKNOLOGI = ['Next.js', 'React', 'TypeScript', 'Tailwind CSS', 'Go (Gin)', 'PostgreSQL', 'JWT Auth']

export const metadata = { title: 'Tentang — SIMTAS FILKOM' }

export default function AboutPage() {
  return (
    <div>
      <section className="border-b border-[var(--st-stroke)]">
        <div className="landing-container max-w-3xl pb-16 pt-36 md:pt-44">
          <Reveal>
            <span className="landing-eyebrow">Tentang Sistem</span>
          </Reveal>
          <Reveal delay={80}>
            <h1 className="landing-display mt-5 text-4xl md:text-5xl">Apa itu SIMTAS FILKOM?</h1>
          </Reveal>
          <Reveal delay={160}>
            <p className="mt-6 text-base leading-relaxed text-[var(--st-muted)] md:text-lg">
              <span className="text-[var(--st-text)]">SIMTAS FILKOM</span> (Sistem Manajemen Tugas Akhir
              Skripsi) adalah platform digital milik Fakultas Ilmu Komputer Universitas Djuanda yang
              mengelola seluruh proses Tugas Akhir Skripsi mahasiswa dalam satu ekosistem yang
              terpadu, transparan, dan aman.
            </p>
          </Reveal>
        </div>
      </section>

      <section className="border-b border-[var(--st-stroke)]">
        <div className="landing-container grid gap-12 py-16 lg:grid-cols-2">
          <Reveal>
            <h2 className="flex items-center gap-2 text-lg font-medium text-[var(--st-text)]">
              <Target className="h-5 w-5 text-[var(--st-accent-from)]" /> Tujuan
            </h2>
            <ul className="mt-5 space-y-3">
              {TUJUAN.map((t) => (
                <li key={t} className="flex gap-3 text-[var(--st-muted)]">
                  <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-[var(--st-accent-from)]" />
                  <span>{t}</span>
                </li>
              ))}
            </ul>
          </Reveal>
          <Reveal delay={80}>
            <h2 className="flex items-center gap-2 text-lg font-medium text-[var(--st-text)]">
              <GraduationCap className="h-5 w-5 text-[var(--st-accent-from)]" /> Manfaat
            </h2>
            <ul className="mt-5 space-y-3">
              {MANFAAT.map((m) => (
                <li key={m} className="flex gap-3 text-[var(--st-muted)]">
                  <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-[var(--st-accent-from)]" />
                  <span>{m}</span>
                </li>
              ))}
            </ul>
          </Reveal>
        </div>
      </section>

      <section className="landing-surface border-b border-[var(--st-stroke)]">
        <div className="landing-container py-16">
          <Reveal>
            <h2 className="flex items-center gap-2 text-lg font-medium text-[var(--st-text)]">
              <Users className="h-5 w-5 text-[var(--st-accent-from)]" /> Siapa yang menggunakan?
            </h2>
          </Reveal>
          <div className="mt-8 grid gap-4 sm:grid-cols-2">
            {PENGGUNA.map((p, i) => {
              const Icon = p.icon
              return (
                <Reveal key={p.role} delay={i * 60} className="h-full">
                  <div className="st-card flex h-full items-start gap-4 rounded-2xl p-6">
                    <span className="flex h-10 w-10 shrink-0 items-center justify-center rounded-xl border border-[var(--st-stroke)] bg-[var(--st-bg)] text-[var(--st-accent-from)]">
                      <Icon className="h-5 w-5" />
                    </span>
                    <div>
                      <h3 className="font-medium text-[var(--st-text)]">{p.role}</h3>
                      <p className="mt-1.5 text-sm text-[var(--st-muted)]">{p.desc}</p>
                    </div>
                  </div>
                </Reveal>
              )
            })}
          </div>
        </div>
      </section>

      <section className="border-b border-[var(--st-stroke)]">
        <div className="landing-container py-16">
          <Reveal>
            <h2 className="flex items-center gap-2 text-lg font-medium text-[var(--st-text)]">
              <Cpu className="h-5 w-5 text-[var(--st-accent-from)]" /> Teknologi yang digunakan
            </h2>
          </Reveal>
          <div className="mt-8 flex flex-wrap gap-3">
            {TEKNOLOGI.map((t) => (
              <span key={t} className="rounded-full border border-[var(--st-stroke)] bg-[var(--st-surface)] px-4 py-1.5 text-sm text-[var(--st-text)]">
                {t}
              </span>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}
