import type { ReactNode } from 'react'
import { cn } from '@/lib/utils/cn'

interface DisplayCardProps {
  className?: string
  step?: string
  icon?: ReactNode
  title?: string
  description?: string
  date?: string
}

export function DisplayCard({
  className,
  step,
  icon,
  title,
  description,
  date,
}: DisplayCardProps) {
  return (
    <div
      className={cn(
        'relative flex h-40 w-88 -skew-y-[8deg] select-none flex-col justify-between rounded-xl border-2 border-st-stroke bg-st-surface/70 px-4 py-3 backdrop-blur-sm transition-all duration-700',
        'hover:border-(--st-accent-from)/40 hover:bg-st-surface',
        '[&>*]:flex [&>*]:items-center [&>*]:gap-2',
        className,
      )}
    >
      <div>
        <span className="relative inline-flex rounded-full bg-(--st-accent-from)/10 p-1">
          {step ? (
            <span className="px-1.5 font-display text-lg font-medium text-(--st-accent-from)">
              {step}
            </span>
          ) : (
            icon
          )}
        </span>
        <p className="text-lg font-medium text-st-text">{title}</p>
      </div>
      <p className="text-lg leading-snug text-st-text">{description}</p>
      {date && <p className="text-st-muted">{date}</p>}
    </div>
  )
}

export default DisplayCard
