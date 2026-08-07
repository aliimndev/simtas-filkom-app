import {
  FileText,
  MessageSquare,
  CheckCircle2,
  Upload,
  CalendarDays,
  Send,
  type LucideIcon,
} from 'lucide-react'
import { cn } from '@/lib/utils/cn'

export interface ActivityItem {
  id: string
  icon?: LucideIcon
  description: string
  actor?: string
  timestamp: string
  type?: 'submission' | 'review' | 'note' | 'upload' | 'schedule' | 'approval'
}

const TYPE_ICONS: Record<string, LucideIcon> = {
  submission: Send,
  review: CheckCircle2,
  note: MessageSquare,
  upload: Upload,
  schedule: CalendarDays,
  approval: CheckCircle2,
}

const TYPE_COLORS: Record<string, string> = {
  submission: 'bg-primary-50 text-primary',
  review: 'bg-success-50 text-success',
  note: 'bg-accent text-accent-foreground',
  upload: 'bg-primary-50 text-primary',
  schedule: 'bg-warning-50 text-warning',
  approval: 'bg-success-50 text-success',
}

export function ActivityList({ items }: { items: ActivityItem[] }) {
  if (items.length === 0) {
    return (
      <div className="py-8 text-center">
        <FileText className="mx-auto h-8 w-8 text-st-muted/40" />
        <p className="mt-2 text-sm text-st-muted">Belum ada aktivitas.</p>
      </div>
    )
  }

  return (
    <div className="divide-y divide-st-stroke">
      {items.map((item) => {
        const Icon = item.icon ?? TYPE_ICONS[item.type ?? ''] ?? FileText
        const colorCls = TYPE_COLORS[item.type ?? ''] ?? 'bg-st-surface-hi text-st-muted'

        return (
          <div
            key={item.id}
            className="flex items-start gap-3 px-0 py-3 first:pt-0 last:pb-0"
          >
            <div
              className={cn(
                'mt-0.5 flex h-7 w-7 shrink-0 items-center justify-center rounded-full',
                colorCls,
              )}
            >
              <Icon className="h-3.5 w-3.5" />
            </div>
            <div className="min-w-0 flex-1">
              <p className="text-sm text-st-text">{item.description}</p>
              <div className="mt-0.5 flex items-center gap-2 text-[11px] text-st-muted">
                {item.actor && <span>{item.actor}</span>}
                {item.actor && <span aria-hidden>·</span>}
                <time>{item.timestamp}</time>
              </div>
            </div>
          </div>
        )
      })}
    </div>
  )
}
