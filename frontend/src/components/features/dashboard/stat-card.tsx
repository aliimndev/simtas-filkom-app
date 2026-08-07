import Link from 'next/link'
import { ArrowUpRight, type LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'

export interface StatCardProps {
  title: string
  value: string | number
  icon: LucideIcon
  href?: string
  suffix?: string
}

export function StatCard({ title, value, icon: Icon, href, suffix }: StatCardProps) {
  const inner = (
    <Card className="h-full transition-colors group-hover:border-primary/40">
      <CardContent className="p-4">
        <div className="flex items-center justify-between gap-2">
          <p className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">{title}</p>
          <Icon className="h-4 w-4 shrink-0 text-muted-foreground/70" aria-hidden />
        </div>
        <p className="mt-2 text-2xl font-semibold tabular-nums tracking-tight text-foreground">
          {value}
          {suffix && <span className="ml-1 text-sm font-normal text-muted-foreground">{suffix}</span>}
        </p>
      </CardContent>
    </Card>
  )

  if (!href) return inner

  return (
    <Link
      href={href}
      className="group block h-full focus-visible:outline-none"
      aria-label={`${title}: ${value}`}
    >
      <div className="h-full rounded-lg transition-colors group-focus-visible:ring-2 group-focus-visible:ring-ring group-focus-visible:ring-offset-2">
        {inner}
      </div>
    </Link>
  )
}

export function ViewAllLink({ href, label = 'Lihat Semua' }: { href: string; label?: string }) {
  return (
    <Link
      href={href}
      className="inline-flex items-center gap-1 text-sm font-medium text-primary-700 transition-colors hover:text-primary-900"
    >
      {label}
      <ArrowUpRight className="h-3.5 w-3.5" />
    </Link>
  )
}
