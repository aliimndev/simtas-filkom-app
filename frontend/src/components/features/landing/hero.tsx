import { ArrowRight } from 'lucide-react'
import { Reveal } from './reveal'
import { RoleCycler } from './role-cycler'
import { LandingButton } from './landing-button'
import { FlickeringGrid } from '@/components/ui/flickering-grid'

export function HeroSection() {
  return (
    <section className="relative overflow-hidden border-b border-st-stroke">
      <FlickeringGrid
        color="#07a2b6"
        maxOpacity={0.35}
        className="pointer-events-none absolute inset-0"
      />
      {/* accent glow */}
      <div
        aria-hidden
        className="pointer-events-none absolute left-1/2 top-[-10%] h-160 w-160 -translate-x-1/2 rounded-full opacity-20 blur-[120px]"
        style={{ background: 'radial-gradient(circle, var(--st-accent-to), transparent 60%)' }}
      />

      <div className="landing-container relative flex flex-col items-center pt-36 pb-24 text-center md:pt-44 md:pb-32">
        {/* Eyebrow */}
        <Reveal>
          <span className="landing-eyebrow">
            SIMTAS://FILKOM <span className="text-st-stroke">·</span> Universitas Djuanda
          </span>
        </Reveal>

        {/* Headline */}
        <Reveal delay={80}>
          <h1 className="landing-display mt-7 max-w-4xl text-[2.6rem] leading-[0.98] sm:text-6xl md:text-7xl">
            Satu sistem untuk seluruh perjalanan{' '}
            <span className="accent-text italic">Tugas Akhir Skripsi</span>.
          </h1>
        </Reveal>

        {/* Role Cycler */}
        <Reveal delay={160}>
          <p className="mt-7 flex items-center justify-center gap-2 text-base text-st-muted md:text-lg">
            Dibangun untuk <RoleCycler />.
          </p>
        </Reveal>

        {/* Description */}
        <Reveal delay={240}>
          <p className="mt-6 max-w-xl text-sm leading-relaxed text-st-muted md:text-base">
            Dari pengajuan judul, bimbingan, seminar, hingga sidang dan arsip — seluruh proses
            Tugas Akhir Fakultas Ilmu Komputer dikelola dalam satu platform.
          </p>
        </Reveal>

        {/* CTA Buttons */}
        <Reveal delay={320}>
          <div className="mt-10 flex flex-col items-center gap-3 sm:flex-row">
            <LandingButton href="/login">
              Masuk ke Sistem <ArrowRight className="h-4 w-4" />
            </LandingButton>
            <LandingButton href="#alur" variant="outline">
              Lihat Alur
            </LandingButton>
          </div>
        </Reveal>
      </div>

      {/* Scroll indicator */}
      <div className="flex flex-col items-center gap-2 pb-8">
        <span className="font-mono text-[0.65rem] uppercase tracking-[0.25em] text-st-muted">
          scroll
        </span>
        <span className="relative h-10 w-px overflow-hidden bg-st-stroke">
          <span className="accent-gradient absolute inset-x-0 top-0 h-1/3 animate-st-scroll" />
        </span>
      </div>
    </section>
  )
}