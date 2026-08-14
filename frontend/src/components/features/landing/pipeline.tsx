import {
  Archive,
  CalendarCheck,
  FileText,
  MessagesSquare,
  SearchCheck,
  UserCheck,
  type LucideIcon,
} from 'lucide-react'
import { Reveal } from './reveal'

interface Stage {
  step: string
  title: string
  desc: string
  icon: LucideIcon
  span: string
}

const STAGES: Stage[] = [
  {
    step: '01',
    title: 'Pengajuan Judul',
    desc: 'Mahasiswa mengajukan judul Tugas Akhir Skripsi secara digital.',
    icon: FileText,
    span: 'lg:col-span-4',
  },
  {
    step: '02',
    title: 'Review Judul',
    desc: 'Kaprodi menyetujui atau menolak beserta catatan.',
    icon: SearchCheck,
    span: 'lg:col-span-2',
  },
  {
    step: '03',
    title: 'Penugasan Pembimbing',
    desc: 'Dosen pembimbing ditetapkan untuk mendampingi.',
    icon: UserCheck,
    span: 'lg:col-span-2',
  },
  {
    step: '04',
    title: 'Bimbingan & Dokumen',
    desc: 'Log bimbingan tercatat dan dokumen diunggah bertahap.',
    icon: MessagesSquare,
    span: 'lg:col-span-4',
  },
  {
    step: '05',
    title: 'Seminar & Sidang',
    desc: 'Seminar proposal hingga sidang akhir terjadwal.',
    icon: CalendarCheck,
    span: 'lg:col-span-3',
  },
  {
    step: '06',
    title: 'Arsip & Kelulusan',
    desc: 'Dokumen final diarsipkan dan status kelulusan dicatat.',
    icon: Archive,
    span: 'lg:col-span-3',
  },
]

export function PipelineSection() {
  return (
    <section id="alur" className="relative overflow-hidden">
      {/* Soft accent wash so the section still opens on a faint teal breath
          after the feather, without a hard edge. */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-x-0 top-0 h-80 bg-linear-to-b from-transparent to-(--st-accent-from) opacity-[0.04]"
      />
      <div className="landing-container py-20 md:py-28">
        <Reveal>
          <span className="landing-eyebrow">
            <span className="h-px w-8 bg-st-stroke" /> Alur Tugas Akhir
          </span>
        </Reveal>
        <Reveal delay={80}>
          <h2 className="landing-heading mt-5 max-w-2xl text-3xl md:text-5xl">
            Enam tahap, satu <span className="accent-text italic">jejak</span> utuh.
          </h2>
        </Reveal>
        <Reveal delay={160}>
          <p className="mt-4 max-w-xl text-sm leading-relaxed text-st-muted md:text-base">
            Setiap mahasiswa melalui urutan yang sama—urutan ini bukan dekorasi, melainkan proses
            nyata yang menentukan maju atau tertunda.
          </p>
        </Reveal>

        <Reveal delay={220}>
          <div className="mt-16 grid grid-cols-1 gap-px overflow-hidden rounded-2xl border border-st-stroke bg-st-stroke sm:grid-cols-2 lg:grid-cols-6">
            {STAGES.map((stage) => {
              const Icon = stage.icon
              return (
                <article
                  key={stage.step}
                  className={`relative flex min-h-44 flex-col overflow-hidden bg-st-surface p-6 transition-colors hover:bg-st-surface-hi md:min-h-52 md:p-8 ${stage.span}`}
                >
                  <div
                    aria-hidden
                    className="pointer-events-none absolute -right-16 -top-16 h-40 w-40 rounded-full bg-(--st-accent-from)/10 blur-2xl"
                  />
                  <div className="relative flex items-start justify-between gap-4">
                    <span className="accent-ring flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-st-surface-hi">
                      <Icon className="h-6 w-6 text-(--st-accent-from)" aria-hidden />
                    </span>
                    <span
                      aria-hidden
                      className="select-none font-display text-5xl font-semibold tracking-tight text-st-text opacity-[0.07] md:text-6xl"
                    >
                      {stage.step}
                    </span>
                  </div>
                  <div className="relative mt-6 md:mt-8">
                    <h3 className="text-xl font-semibold tracking-tight text-st-text md:text-2xl">
                      {stage.title}
                    </h3>
                    <p className="mt-3 max-w-2xl text-sm leading-7 text-st-muted md:text-base">
                      {stage.desc}
                    </p>
                  </div>
                </article>
              )
            })}
          </div>
        </Reveal>
      </div>
    </section>
  )
}
