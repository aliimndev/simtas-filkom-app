import { ArrowRight } from 'lucide-react'
import { Reveal } from './reveal'
import { LandingButton } from './landing-button'
import { Parallax } from './parallax'

export function CtaSection() {
  return (
    <section className="relative overflow-hidden">
      <Parallax
        speed={0.5}
        className="pointer-events-none absolute inset-0 flex items-center justify-center"
      >
        <div className="h-160 w-160 rounded-full bg-(--st-accent-to) opacity-[0.09] blur-[120px]" />
      </Parallax>
      <div className="landing-container flex flex-col items-center py-24 text-center md:py-32">
        <Reveal>
          <h2 className="landing-heading mx-auto max-w-2xl text-3xl md:text-5xl">
            Lanjutkan proses <span className="accent-text italic">Tugas Akhir Anda</span>.
          </h2>
        </Reveal>
        <Reveal delay={80}>
          <p className="mx-auto mt-5 max-w-xl text-sm leading-relaxed text-st-muted md:text-base">
            Akses SIMTAS FILKOM untuk mengelola proses Tugas Akhir secara terstruktur dan
            terdokumentasi.
          </p>
        </Reveal>
        <Reveal delay={160}>
          <div className="mt-9 flex flex-col items-center gap-3 sm:flex-row">
            <LandingButton href="/login">
              Masuk ke Sistem <ArrowRight className="h-4 w-4" />
            </LandingButton>
          </div>
        </Reveal>
      </div>
    </section>
  )
}
