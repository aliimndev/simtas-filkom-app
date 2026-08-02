import Link from 'next/link'
import { BookOpen, GraduationCap, ClipboardCheck, ArrowRight } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'

const MAHASISWA = [
  { title: 'Cara Login', desc: 'Masuk dengan email dan password yang diberikan administrator.' },
  { title: 'Ajukan Judul', desc: 'Buka Skripsi Saya → Ajukan Judul, lengkapi formulir, lalu kirim.' },
  { title: 'Upload Proposal', desc: 'Lewat menu Dokumen, unggah file proposal/skripsi sesuai ketentuan.' },
  { title: 'Bimbingan', desc: 'Catat log bimbingan dan menunggu validasi dosen pembimbing.' },
  { title: 'Seminar & Sidang', desc: 'Ikuti jadwal seminar dan sidang yang ditetapkan fakultas.' },
]

const DOSEN = [
  { title: 'Review Proposal', desc: 'Buka daftar dokumen dan berikan review/penilaian.' },
  { title: 'Validasi Bimbingan', desc: 'Setujui log bimbingan mahasiswa yang Anda bimbing.' },
  { title: 'Penilaian', desc: 'Berikan nilai di seminar dan sidang melalui menu terkait.' },
]

const KAPRODI = [
  { title: 'Approval Judul', desc: 'Tinjau dan setujui/tolak pengajuan judul mahasiswa.' },
  { title: 'Monitoring', desc: 'Pantau progres skripsi mahasiswa lewat dashboard.' },
  { title: 'Penjadwalan', desc: 'Atur jadwal seminar dan sidang beserta pengujinya.' },
]

function Section({ type }: { type: 'mahasiswa' | 'dosen' | 'kaprodi' }) {
  const data =
    type === 'mahasiswa'
      ? { title: 'Panduan untuk Mahasiswa', icon: GraduationCap, items: MAHASISWA }
      : type === 'dosen'
        ? { title: 'Panduan untuk Dosen', icon: ClipboardCheck, items: DOSEN }
        : { title: 'Panduan untuk Kaprodi', icon: BookOpen, items: KAPRODI }
  const Icon = data.icon
  return (
    <div>
      <h2 className="flex items-center gap-2 text-xl font-semibold text-foreground">
        <Icon className="h-5 w-5 text-primary" /> {data.title}
      </h2>
      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        {data.items.map((item, i) => (
          <Card key={item.title}>
            <CardContent className="p-5">
              <div className="flex items-center gap-3">
                <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md bg-primary-50 text-xs font-bold text-primary">
                  {i + 1}
                </span>
                <h3 className="font-semibold text-foreground">{item.title}</h3>
              </div>
              <p className="mt-2 text-sm text-muted-foreground">{item.desc}</p>
            </CardContent>
          </Card>
        ))}
      </div>
    </div>
  )
}

export const metadata = { title: 'Panduan — SIMTAS FILKOM' }

export default function GuidePage() {
  return (
    <div>
      <section className="border-b border-border">
        <div className="landing-container max-w-3xl py-16">
          <span className="landing-eyebrow">
            <BookOpen className="h-4 w-4" /> Panduan Penggunaan
          </span>
          <h1 className="landing-display mt-4 text-4xl md:text-5xl">Panduan untuk setiap peran</h1>
          <p className="mt-5 text-lg text-muted-foreground">
            Ikuti panduan sesuai peran Anda untuk memanfaatkan seluruh fitur sistem dengan benar.
          </p>
        </div>
      </section>

      <section className="landing-surface border-b border-border">
        <div className="landing-container space-y-14 py-16">
          <Section type="mahasiswa" />
          <Section type="dosen" />
          <Section type="kaprodi" />
        </div>
      </section>

      <section className="border-b border-border">
        <div className="landing-container py-16 text-center">
          <h2 className="landing-display text-2xl md:text-3xl">Butuh bantuan lebih lanjut?</h2>
          <p className="mx-auto mt-3 max-w-xl text-muted-foreground">
            Kunjungi FAQ untuk pertanyaan umum atau hubungi kami untuk bantuan langsung.
          </p>
          <div className="mt-7 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link
              href="/faq"
              className="inline-flex items-center gap-2 rounded-lg border border-border px-5 py-2.5 text-sm font-medium text-foreground transition hover:bg-muted"
            >
              Buka FAQ
            </Link>
            <Link
              href="/contact"
              className="inline-flex items-center gap-2 rounded-lg bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground transition hover:bg-primary-700"
            >
              Hubungi Kami <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>
    </div>
  )
}
