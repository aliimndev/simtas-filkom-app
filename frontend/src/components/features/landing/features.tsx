import { Archive, Bell, CalendarCheck, FileText, LayoutDashboard, MessagesSquare, type LucideIcon } from 'lucide-react'
import { Reveal } from './reveal'

interface Feature {
  icon: LucideIcon
  title: string
  desc: string
}

const FEATURES: Feature[] = [
  { icon: FileText, title: 'Pengajuan Judul', desc: 'Ajukan judul dan dapatkan review dari kaprodi secara digital.' },
  { icon: MessagesSquare, title: 'Bimbingan Online', desc: 'Catat log bimbingan dan menunggu validasi dosen pembimbing.' },
  { icon: CalendarCheck, title: 'Seminar & Sidang', desc: 'Penjadwalan, penguji, dan penilaian dalam satu alur.' },
  { icon: Archive, title: 'Arsip Digital', desc: 'Dokumen final terdokumentasi dan mudah diakses kembali.' },
  { icon: LayoutDashboard, title: 'Dashboard Real-time', desc: 'Statistik dan status per peran: mahasiswa, dosen, kaprodi, admin.' },
  { icon: Bell, title: 'Notifikasi Email', desc: 'Pemberitahuan penting terkirim ke pemangku kepentingan.' },
]

export function FeaturesSection() {
  return (
    <section className="landing-surface border-b border-st-stroke">
      <div className="landing-container py-20 md:py-28">
        <Reveal>
          <span className="landing-eyebrow">
            <span className="h-px w-8 bg-st-stroke" /> Kemampuan
          </span>
        </Reveal>
        <Reveal delay={80}>
          <h2 className="landing-display mt-5 text-3xl md:text-5xl">
            Yang bisa Anda <span className="accent-text italic">lakukan</span>.
          </h2>
        </Reveal>

        <div className="mt-14 grid gap-4 md:grid-cols-2 lg:grid-cols-3">
          {FEATURES.map((f, i) => {
            const Icon = f.icon
            return (
              <Reveal key={f.title} delay={i * 60} className="h-full">
                <div className="group st-card flex h-full flex-col gap-4 rounded-2xl p-6 transition-colors hover:border-(--st-accent-from)/40 md:p-7">
                  <span className="flex h-10 w-10 items-center justify-center rounded-xl border border-st-stroke bg-st-bg text-(--st-accent-from) transition-colors group-hover:text-(--st-accent-to)">
                    <Icon className="h-5 w-5" />
                  </span>
                  <h3 className="text-base font-medium text-st-text">{f.title}</h3>
                  <p className="text-sm leading-relaxed text-st-muted">{f.desc}</p>
                </div>
              </Reveal>
            )
          })}
        </div>
      </div>
    </section>
  )
}
