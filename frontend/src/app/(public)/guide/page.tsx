import Link from 'next/link'
import { BookOpen, GraduationCap, ClipboardCheck, ArrowRight } from 'lucide-react'
import { Reveal } from '@/components/features/landing/reveal'

const MAHASISWA = [
  { title: 'Cara Login', desc: 'Masuk dengan email dan password yang diberikan administrator.' },
  { title: 'Ajukan Judul', desc: 'Buka Tugas Akhir Skripsi → Ajukan Judul, lengkapi formulir, lalu kirim.' },
  { title: 'Unggah Dokumen', desc: 'Lewat menu Dokumen, unggah proposal/dokumen sesuai ketentuan.' },
  { title: 'Bimbingan', desc: 'Catat log bimbingan dan menunggu validasi dosen pembimbing.' },
  { title: 'Seminar & Sidang', desc: 'Ikuti jadwal seminar dan sidang yang ditetapkan fakultas.' },
]

const DOSEN = [
  { title: 'Review Dokumen', desc: 'Buka daftar dokumen dan berikan review/penilaian.' },
  { title: 'Validasi Bimbingan', desc: 'Setujui log bimbingan mahasiswa yang Anda bimbing.' },
  { title: 'Penilaian', desc: 'Berikan nilai di seminar dan sidang melalui menu terkait.' },
]

const KAPRODI = [
  { title: 'Approval Judul', desc: 'Tinjau dan setujui/tolak pengajuan judul mahasiswa.' },
  { title: 'Monitoring', desc: 'Pantau progres Tugas Akhir mahasiswa lewat dashboard.' },
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
      <h2 className="flex items-center gap-2 text-lg font-medium text-st-text">
        <Icon className="h-5 w-5 text-(--st-accent-from)" /> {data.title}
      </h2>
      <div className="mt-6 grid gap-4 sm:grid-cols-2">
        {data.items.map((item, i) => (
          <div key={item.title} className="st-card rounded-2xl p-5">
            <div className="flex items-center gap-3">
              <span className="flex h-7 w-7 shrink-0 items-center justify-center rounded-md accent-gradient text-xs font-bold text-st-bg">
                {i + 1}
              </span>
              <h3 className="font-medium text-st-text">{item.title}</h3>
            </div>
            <p className="mt-2 text-sm text-st-muted">{item.desc}</p>
          </div>
        ))}
      </div>
    </div>
  )
}

export const metadata = { title: 'Panduan — SIMTAS FILKOM' }

export default function GuidePage() {
  return (
    <div>
      <section className="border-b border-st-stroke">
        <div className="landing-container max-w-3xl pb-16 pt-36 md:pt-44">
          <Reveal><span className="landing-eyebrow">Panduan Penggunaan</span></Reveal>
          <Reveal delay={80}><h1 className="landing-display mt-5 text-4xl md:text-5xl">Panduan untuk setiap peran</h1></Reveal>
          <Reveal delay={160}>
            <p className="mt-5 text-base text-st-muted md:text-lg">
              Ikuti panduan sesuai peran Anda untuk memanfaatkan seluruh fitur sistem dengan benar.
            </p>
          </Reveal>
        </div>
      </section>

      <section className="landing-surface border-b border-st-stroke">
        <div className="landing-container space-y-14 py-16">
          <Reveal><Section type="mahasiswa" /></Reveal>
          <Reveal delay={80}><Section type="dosen" /></Reveal>
          <Reveal delay={160}><Section type="kaprodi" /></Reveal>
        </div>
      </section>

      <section className="border-b border-st-stroke">
        <div className="landing-container py-16 text-center">
          <h2 className="landing-display text-2xl md:text-3xl">Butuh bantuan lebih lanjut?</h2>
          <p className="mx-auto mt-3 max-w-xl text-st-muted">
            Kunjungi FAQ untuk pertanyaan umum atau hubungi kami untuk bantuan langsung.
          </p>
          <div className="mt-7 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link href="/faq" className="accent-ring inline-flex items-center gap-2 rounded-full border border-st-stroke bg-st-surface px-5 py-2.5 text-sm font-medium text-st-text transition hover:border-(--st-accent-from)/40 hover:bg-st-surface-hi">
              Buka FAQ
            </Link>
            <Link href="/contact" className="accent-ring inline-flex items-center gap-2 rounded-full bg-st-text px-5 py-2.5 text-sm font-medium text-st-bg transition hover:opacity-90">
              Hubungi Kami <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>
    </div>
  )
}
