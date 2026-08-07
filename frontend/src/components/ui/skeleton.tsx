import { cn } from '@/lib/utils/cn'

export function Skeleton({ className, ...props }: React.HTMLAttributes<HTMLDivElement>) {
  return <div className={cn('animate-pulse rounded-lg bg-muted', className)} {...props} />
}

// ponytail: single card-row placeholder for every list page; good enough for table/grid pages too
export function ListSkeleton({ count = 5, label = 'Memuat data…' }: { count?: number; label?: string }) {
  return (
    <div className="space-y-3" role="status" aria-label={label}>
      {Array.from({ length: count }).map((_, i) => (
        <div key={i} className="flex items-center justify-between gap-3 rounded-xl border border-border bg-background p-5">
          <div className="flex min-w-0 flex-1 items-center gap-3">
            <Skeleton className="h-11 w-11 shrink-0 rounded-xl" />
            <div className="min-w-0 flex-1 space-y-2">
              <Skeleton className="h-4 w-2/5 max-w-52" />
              <Skeleton className="h-3 w-full max-w-xl" />
              <Skeleton className="h-3 w-3/5 max-w-sm" />
            </div>
          </div>
          <Skeleton className="hidden h-8 w-24 shrink-0 md:block" />
        </div>
      ))}
    </div>
  )
}
