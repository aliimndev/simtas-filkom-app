import { formatDate } from '@/lib/utils/date'
import { Reveal } from '@/components/features/landing/reveal'

function greeting(): string {
  const h = new Date().getHours()
  if (h < 11) return 'selamat pagi'
  if (h < 15) return 'selamat siang'
  if (h < 18) return 'selamat sore'
  return 'selamat malam'
}

export function DashboardHeader({ name, subtitle }: { name: string; subtitle?: string }) {
  const firstName = name.split(' ')[0]
  return (
    <Reveal>
      <div className="flex flex-wrap items-end justify-between gap-x-6 gap-y-3">
        <div>
          <p className="landing-eyebrow">
            Sistem Manajemen Tugas Akhir
          </p>
          <h1 className="mt-3 text-balance landing-heading text-3xl sm:text-4xl">
            <span className="accent-text italic">{firstName}</span>, {greeting()}.
          </h1>
          {subtitle && <p className="mt-2 text-sm text-st-muted">{subtitle}</p>}
        </div>
        <p className="font-mono text-[0.7rem] uppercase tracking-[0.25em] tabular-nums text-st-muted" aria-hidden>
          {formatDate(new Date())}
        </p>
      </div>
    </Reveal>
  )
}

