import { ArrowRight } from 'lucide-react'
import { Reveal } from './reveal'

interface Stage {
  step: string
  title: string
  desc: string
}

const STAGES: Stage[] = [
  { step: '01', title: 'Pengajuan Judul', desc: 'Mahasiswa mengajukan judul Tugas Akhir Skripsi secara digital.' },
  { step: '02', title: 'Review Judul', desc: 'Kaprodi menyetujui atau menolak beserta catatan.' },
  { step: '03', title: 'Penugasan Pembimbing', desc: 'Dosen pembimbing ditetapkan untuk mendampingi.' },
  { step: '04', title: 'Bimbingan & Dokumen', desc: 'Log bimbingan tercatat dan dokumen diunggah bertahap.' },
  { step: '05', title: 'Seminar & Sidang', desc: 'Seminar proposal hingga sidang akhir terjadwal.' },
  { step: '06', title: 'Arsip & Kelulusan', desc: 'Dokumen final diarsipkan dan status kelulusan dicatat.' },
]

export function PipelineSection() {
  return (
    <section id="alur" className="border-b border-st-stroke">
      <div className="landing-container py-20 md:py-28">
        <Reveal>
          <span className="landing-eyebrow">
            <span className="h-px w-8 bg-st-stroke" /> Alur Tugas Akhir
          </span>
        </Reveal>
        <Reveal delay={80}>
          <h2 className="landing-display mt-5 max-w-2xl text-3xl md:text-5xl">
            Enam tahap, satu <span className="accent-text italic">jejak</span> utuh.
          </h2>
        </Reveal>
        <Reveal delay={160}>
          <p className="mt-4 max-w-xl text-sm leading-relaxed text-st-muted md:text-base">
            Setiap mahasiswa melalui urutan yang sama—urutan ini bukan dekorasi, melainkan proses
            nyata yang menentukan maju atau tertunda.
          </p>
        </Reveal>

        <div className="mt-14 grid gap-px overflow-hidden rounded-2xl border border-st-stroke bg-st-stroke sm:grid-cols-2 lg:grid-cols-3">
          {STAGES.map((stage, i) => (
            <Reveal key={stage.step} delay={i * 60} className="h-full">
              <div className="group relative flex h-full flex-col gap-3 bg-st-surface p-6 transition-colors hover:bg-st-surface-hi md:p-8">
                <div className="flex items-baseline justify-between">
                  <span className="font-display text-4xl text-st-muted transition-colors group-hover:text-st-text md:text-5xl">
                    {stage.step}
                  </span>
                  {i < STAGES.length - 1 && (
                    <ArrowRight className="h-4 w-4 text-st-stroke transition-colors group-hover:text-(--st-accent-from)" />
                  )}
                </div>
                <h3 className="text-base font-medium text-st-text md:text-lg">{stage.title}</h3>
                <p className="text-sm leading-relaxed text-st-muted">{stage.desc}</p>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  )
}
