import { GraduationCap, Target, Users, Cpu, CheckCircle2, Sparkles } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'

const TUJUAN = [
  'Mempermudah administrasi tugas akhir & skripsi yang sebelumnya manual.',
  'Menjamin transparansi proses dari pengajuan judul hingga kelulusan.',
  'Mendigitalisasi arsip dan dokumen akademik agar mudah diakses kembali.',
]

const MANFAAT = [
  'Mahasiswa dapat memantau status skripsi secara real-time.',
  'Dosen pembimbing dan penguji mendapat alur validasi yang jelas.',
  'Kaprodi dan admin memperoleh laporan serta audit yang ringkas.',
  'Mengurangi dokumen fisik dan mempercepat alur persetujuan.',
]

const PENGGUNA = [
  { role: 'Mahasiswa', desc: 'Mengajukan judul, mengunggah dokumen, dan memantau bimbingan.' },
  { role: 'Dosen Pembimbing & Penguji', desc: 'Memvalidasi bimbingan, mereview dokumen, dan melakukan penilaian.' },
  { role: 'Kaprodi', desc: 'Menyetujui judul, memantau progres, dan mengelola penjadwalan.' },
  { role: 'Administrator Fakultas', desc: 'Mengelola pengguna, konfigurasi sistem, dan data master.' },
]

const TEKNOLOGI = ['Next.js', 'React', 'TypeScript', 'Tailwind CSS', 'Go (Gin)', 'PostgreSQL', 'JWT Auth']

export const metadata = { title: 'Tentang — SIMTAS FILKOM' }

export default function AboutPage() {
  return (
    <div>
      <section className="border-b border-border">
        <div className="landing-container max-w-3xl py-20">
          <span className="landing-eyebrow">
            <Sparkles className="h-4 w-4" /> Tentang Sistem
          </span>
          <h1 className="landing-display mt-4 text-4xl md:text-5xl">Apa itu SIMTAS FILKOM?</h1>
          <p className="mt-6 text-lg leading-relaxed text-muted-foreground">
            <strong className="text-foreground">SIMTAS FILKOM</strong> (Sistem Informasi Manajemen Tugas Akhir dan
            Skripsi) adalah platform digital milik Fakultas Ilmu Komputer Universitas Djuanda yang
            mengelola seluruh proses tugas akhir dan skripsi mahasiswa dalam satu ekosistem yang
            terpadu, transparan, dan aman.
          </p>
        </div>
      </section>

      <section className="border-b border-border">
        <div className="landing-container py-16">
          <div className="grid gap-10 lg:grid-cols-2">
            <div>
              <h2 className="flex items-center gap-2 text-xl font-semibold text-foreground">
                <Target className="h-5 w-5 text-primary" /> Tujuan
              </h2>
              <ul className="mt-5 space-y-3">
                {TUJUAN.map((t) => (
                  <li key={t} className="flex gap-3 text-muted-foreground">
                    <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-success" />
                    <span>{t}</span>
                  </li>
                ))}
              </ul>
            </div>
            <div>
              <h2 className="flex items-center gap-2 text-xl font-semibold text-foreground">
                <GraduationCap className="h-5 w-5 text-primary" /> Manfaat
              </h2>
              <ul className="mt-5 space-y-3">
                {MANFAAT.map((m) => (
                  <li key={m} className="flex gap-3 text-muted-foreground">
                    <CheckCircle2 className="mt-0.5 h-5 w-5 shrink-0 text-success" />
                    <span>{m}</span>
                  </li>
                ))}
              </ul>
            </div>
          </div>
        </div>
      </section>

      <section className="landing-surface border-b border-border">
        <div className="landing-container py-16">
          <h2 className="flex items-center gap-2 text-xl font-semibold text-foreground">
            <Users className="h-5 w-5 text-primary" /> Siapa yang menggunakan?
          </h2>
          <div className="mt-8 grid gap-5 sm:grid-cols-2">
            {PENGGUNA.map((p) => (
              <Card key={p.role}>
                <CardContent className="p-6">
                  <h3 className="font-semibold text-foreground">{p.role}</h3>
                  <p className="mt-1.5 text-sm text-muted-foreground">{p.desc}</p>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      <section className="border-b border-border">
        <div className="landing-container py-16">
          <h2 className="flex items-center gap-2 text-xl font-semibold text-foreground">
            <Cpu className="h-5 w-5 text-primary" /> Teknologi yang digunakan
          </h2>
          <div className="mt-8 flex flex-wrap gap-3">
            {TEKNOLOGI.map((t) => (
              <span key={t} className="rounded-full border border-border bg-background px-4 py-1.5 text-sm text-foreground">
                {t}
              </span>
            ))}
          </div>
        </div>
      </section>
    </div>
  )
}
