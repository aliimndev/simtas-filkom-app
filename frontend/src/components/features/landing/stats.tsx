import { Reveal } from './reveal'

const NARRATIVE = [
  {
    id: '01',
    label: 'Problem',
    text: 'Proses yang melibatkan banyak tahapan.',
  },
  {
    id: '02',
    label: 'Solution',
    text: 'Satu sistem untuk menghubungkan seluruh proses.',
  },
  {
    id: '03',
    label: 'Result',
    text: 'Proses lebih mudah dipantau dan terdokumentasi.',
  },
]

export function ProblemSolutionSection() {
  return (
    <section className="relative overflow-hidden border-b border-st-stroke">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_top,var(--st-accent-from),transparent_55%)] opacity-[0.08]"
      />
      <div className="landing-container relative py-20 md:py-28">
        <Reveal>
          <span className="landing-eyebrow">SATU ALUR, LEBIH TERARAH</span>
        </Reveal>
        <Reveal delay={80}>
          <h2 className="landing-display mt-5 max-w-3xl text-3xl md:text-5xl">
            Proses Tugas Akhir tidak harus berjalan sendiri-sendiri.
          </h2>
        </Reveal>
        <Reveal delay={140}>
          <p className="mt-5 max-w-3xl text-sm leading-relaxed text-st-muted md:text-base">
            Pengajuan, review, bimbingan, seminar, sidang, dan arsip melibatkan berbagai tahapan
            dan pihak. SIMTAS menyatukan seluruh proses tersebut dalam satu alur digital yang
            terstruktur.
          </p>
        </Reveal>

        <div className="mt-14 grid gap-6 md:grid-cols-3">
          {NARRATIVE.map((item, index) => (
            <Reveal key={item.id} delay={index * 90}>
              <article className="group relative overflow-hidden rounded-2xl border border-st-stroke bg-st-surface/80 p-6 backdrop-blur-sm transition-colors duration-300 hover:bg-st-surface-hi">
                <div
                  aria-hidden
                  className="pointer-events-none absolute inset-y-0 left-0 w-px bg-linear-to-b from-transparent via-(--st-accent-from) to-transparent opacity-60"
                />
                <p className="font-mono text-xs tracking-[0.2em] text-st-muted uppercase">
                  {item.id} · {item.label}
                </p>
                <p className="mt-4 text-base leading-relaxed text-st-text md:text-lg">{item.text}</p>
              </article>
            </Reveal>
          ))}
        </div>

        <div
          aria-hidden
          className="pointer-events-none mt-10 h-16 w-full rounded-full bg-linear-to-r from-transparent via-(--st-accent-to) to-transparent opacity-20 blur-2xl"
        />
      </div>
    </section>
  )
}
