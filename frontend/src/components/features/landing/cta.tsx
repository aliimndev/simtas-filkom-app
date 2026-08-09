import { ArrowRight } from 'lucide-react'
import { Reveal } from './reveal'
import { LandingButton } from './landing-button'

const MARQUEE_WORDS = ['PENGAJUAN', 'BIMBINGAN', 'SEMINAR', 'SIDANG', 'ARSIP']

export function CtaSection() {
  return (
    <section className="relative overflow-hidden">
      <div className="landing-container flex flex-col items-center py-24 text-center md:py-32">
        <Reveal>
          <h2 className="landing-display mx-auto max-w-2xl text-3xl md:text-5xl">
            Siap memulai <span className="accent-text italic">Tugas Akhir</span> Anda?
          </h2>
        </Reveal>
        <Reveal delay={80}>
          <p className="mx-auto mt-5 max-w-xl text-sm leading-relaxed text-st-muted md:text-base">
            Masuk untuk mengajukan judul, melanjutkan bimbingan, atau memantau proses.
          </p>
        </Reveal>
        <Reveal delay={160}>
          <div className="mt-9 flex flex-col items-center gap-3 sm:flex-row">
            <LandingButton href="/login">
              Masuk ke Sistem <ArrowRight className="h-4 w-4" />
            </LandingButton>
            <LandingButton href="/faq" variant="outline">
              Baca FAQ
            </LandingButton>
          </div>
        </Reveal>
      </div>

      {/* marquee */}
      <div className="border-t border-st-stroke py-5">
        <div className="flex w-max animate-st-marquee whitespace-nowrap">
          {Array.from({ length: 2 }).map((_, dup) => (
            <div key={dup} className="flex items-center gap-8 pr-8" aria-hidden={dup === 1}>
              {MARQUEE_WORDS.map((w) => (
                <span key={w} className="font-display text-2xl italic text-st-muted md:text-3xl">
                  {w} <span className="text-(--st-accent-from)">•</span>
                </span>
              ))}
            </div>
          ))}
        </div>
      </div>
    </section>
  )
}
