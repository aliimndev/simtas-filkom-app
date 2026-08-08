import Link from 'next/link'
import { ArrowUpRight, type LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils/cn'

export type StatTone = 'primary' | 'secondary' | 'success' | 'warning' | 'danger'

const TONE_CHIP: Record<StatTone, string> = {
  primary: 'bg-primary-50 text-primary',
  secondary: 'bg-secondary-50 text-secondary',
  success: 'bg-success-50 text-success',
  warning: 'bg-warning-50 text-warning',
  danger: 'bg-danger-50 text-danger',
}

export interface StatCardProps {
  title: string
  value: string | number
  icon: LucideIcon
  href?: string
  suffix?: string
  tone?: StatTone
}

export function StatCard({ title, value, icon: Icon, href, suffix, tone = 'primary' }: StatCardProps) {
  const inner = (
    <Card className="h-full border-st-stroke bg-st-surface transition-colors group-hover:border-(--st-accent-from)/40">
      <CardContent className="p-5">
        <div className="flex items-center gap-3">
          <div className={cn('flex h-9 w-9 shrink-0 items-center justify-center rounded-full', TONE_CHIP[tone])}>
            <Icon className="h-4.5 w-4.5" aria-hidden />
          </div>
          <p className="font-mono text-[0.7rem] uppercase leading-tight tracking-[0.2em] text-st-muted">{title}</p>
        </div>
        <p className="mt-3 font-display text-3xl leading-none tabular-nums tracking-tight text-st-text">
          {value}
          {suffix && <span className="ml-1 font-body text-sm font-normal text-st-muted">{suffix}</span>}
        </p>
      </CardContent>
    </Card>
  )

  if (!href) return inner

  return (
    <Link
      href={href}
      className="group block h-full rounded-2xl focus-visible:outline-none"
      aria-label={`${title}: ${value}`}
    >
      <div className="h-full rounded-2xl transition-shadow group-focus-visible:ring-2 group-focus-visible:ring-ring group-focus-visible:ring-offset-2">
        {inner}
      </div>
    </Link>
  )
}

export function ViewAllLink({ href, label = 'Lihat Semua' }: { href: string; label?: string }) {
  return (
    <Link
      href={href}
      className="inline-flex items-center gap-1 font-mono text-[0.7rem] uppercase tracking-[0.2em] text-(--st-accent-to) transition-colors hover:text-(--st-accent-from)"
    >
      {label}
      <ArrowUpRight className="h-3.5 w-3.5" />
    </Link>
  )
}

