import { formatDate } from '@/lib/utils/date'

function greeting(): string {
  const h = new Date().getHours()
  if (h < 11) return 'Selamat pagi'
  if (h < 15) return 'Selamat siang'
  if (h < 18) return 'Selamat sore'
  return 'Selamat malam'
}

export function DashboardHeader({ name, subtitle }: { name: string; subtitle?: string }) {
  const firstName = name.split(' ')[0]
  return (
    <div className="flex flex-wrap items-end justify-between gap-x-6 gap-y-2">
      <div>
        <h1 className="text-xl font-semibold tracking-tight text-foreground sm:text-2xl">
          {greeting()}, {firstName}
        </h1>
        {subtitle && <p className="mt-1 text-sm text-muted-foreground">{subtitle}</p>}
      </div>
      <p className="text-sm tabular-nums text-muted-foreground" aria-hidden>
        {formatDate(new Date())}
      </p>
    </div>
  )
}
