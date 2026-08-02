import Link from 'next/link'
import { ArrowUpRight, type LucideIcon } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils/cn'

export interface StatCardProps {
  title: string
  value: string | number
  icon: LucideIcon
  href?: string
  iconClass?: string
  suffix?: string
}

export function StatCard({ title, value, icon: Icon, href, iconClass, suffix }: StatCardProps) {
  const inner = (
    <Card className="group overflow-hidden">
      <CardContent className="flex items-center justify-between p-5">
        <div>
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">{title}</p>
          <p className="mt-1.5 text-2xl font-bold text-foreground">
            {value}
            {suffix && <span className="ml-0.5 text-sm font-normal text-muted-foreground">{suffix}</span>}
          </p>
        </div>
        <div
          className={cn(
            'flex h-11 w-11 items-center justify-center rounded-xl bg-primary-50 text-primary transition-transform group-hover:scale-110',
            iconClass,
          )}
        >
          <Icon className="h-5 w-5" />
        </div>
      </CardContent>
    </Card>
  )

  if (!href) return inner

  return (
    <Link href={href} className="group block">
      {inner}
    </Link>
  )
}

export function ViewAllLink({ href, label = 'Lihat Semua' }: { href: string; label?: string }) {
  return (
    <Link
      href={href}
      className="inline-flex items-center gap-1 text-xs font-medium text-primary hover:text-primary-700 hover:underline"
    >
      {label}
      <ArrowUpRight className="h-3 w-3" />
    </Link>
  )
}
