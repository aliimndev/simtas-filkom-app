import { CalendarDays } from 'lucide-react'
import { cn } from '@/lib/utils/cn'

export interface UpcomingItem {
  id: string
  title: string
  date: string
  time?: string
  type?: 'bimbingan' | 'seminar' | 'sidang'
}

const TYPE_DOT: Record<string, string> = {
  bimbingan: 'bg-primary',
  seminar: 'bg-warning',
  sidang: 'bg-success',
}

export function UpcomingList({ items }: { items: UpcomingItem[] }) {
  if (items.length === 0) {
    return (
      <div className="py-8 text-center">
        <CalendarDays className="mx-auto h-8 w-8 text-st-muted" />
        <p className="mt-2 text-sm text-st-muted">Tidak ada agenda mendatang.</p>
      </div>
    )
  }

  return (
    <div className="divide-y divide-st-stroke">
      {items.map((item) => {
        const dotCls = TYPE_DOT[item.type ?? ''] ?? 'bg-st-muted'
        return (
          <div
            key={item.id}
            className="flex items-center gap-3 py-3 first:pt-0 last:pb-0"
          >
            <span className={cn('h-2 w-2 shrink-0 rounded-full', dotCls)} />
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-st-text">{item.title}</p>
              <p className="mt-0.5 text-[11px] text-st-muted">
                {item.date}
                {item.time && ` · ${item.time}`}
              </p>
            </div>
          </div>
        )
      })}
    </div>
  )
}
