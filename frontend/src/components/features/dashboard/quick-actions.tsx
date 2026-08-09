import Link from 'next/link'
import { ArrowRight, type LucideIcon } from 'lucide-react'
import { cn } from '@/lib/utils/cn'

export interface QuickActionItem {
  label: string
  href: string
  icon: LucideIcon
  variant?: 'primary' | 'secondary'
}

export function QuickActions({ actions }: { actions: QuickActionItem[] }) {
  return (
    <div className="flex flex-wrap gap-2">
      {actions.map((action) => {
        const Icon = action.icon
        const isPrimary = action.variant === 'primary'
        return (
          <Link
            key={action.href}
            href={action.href}
            className={cn(
              'accent-ring inline-flex items-center gap-2 rounded-full px-4 py-2 text-sm font-medium transition duration-150',
              isPrimary
                ? 'bg-st-text text-st-bg hover:opacity-90'
                : 'border border-st-stroke bg-st-surface text-st-text hover:border-(--st-accent-from)/40 hover:bg-st-surface-hi',
            )}
          >
            <Icon className="h-4 w-4" />
            {action.label}
            {isPrimary && <ArrowRight className="h-3.5 w-3.5" />}
          </Link>
        )
      })}
    </div>
  )
}

