import { ArrowRight } from 'lucide-react'
import { Reveal } from './reveal'
import { LandingButton } from './landing-button'

export function CtaSection() {
  return (
    <section className="relative overflow-hidden">
      <div
        aria-hidden
        className="pointer-events-none absolute inset-0 bg-[radial-gradient(ellipse_at_center,var(--st-accent-to),transparent_60%)] opacity-[0.09]"
      />
      <div className="landing-container flex flex-col items-center py-24 text-center md:py-32">
        <Reveal>
          <h2 className="landing-display mx-auto max-w-2xl text-3xl md:text-5xl">
            Lanjutkan proses Tugas Akhir Anda.
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
