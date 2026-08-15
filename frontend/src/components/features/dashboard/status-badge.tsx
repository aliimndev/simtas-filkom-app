import { cn } from '@/lib/utils/cn'

type StatusVariant = 'pending' | 'approved' | 'rejected' | 'draft' | 'in_progress' | 'completed'

const VARIANT_STYLES: Record<StatusVariant, string> = {
  pending: 'bg-warning-50 text-warning border border-warning/20',
  approved: 'bg-success-50 text-success border border-success/20',
  rejected: 'bg-danger-50 text-danger-700 border border-danger/20',
  draft: 'bg-st-surface-hi text-st-muted border border-st-stroke',
  in_progress: 'bg-primary-50 text-primary border border-primary/20',
  completed: 'bg-success-50 text-success border border-success/20',
}

const VARIANT_DOT: Record<StatusVariant, string> = {
  pending: 'bg-warning',
  approved: 'bg-success',
  rejected: 'bg-danger',
  draft: 'bg-muted-foreground',
  in_progress: 'bg-primary',
  completed: 'bg-success',
}

export function StatusBadge({
  variant,
  label,
  className,
}: {
  variant: StatusVariant
  label: string
  className?: string
}) {
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-[11px] font-medium',
        VARIANT_STYLES[variant],
        className,
      )}
    >
      <span className={cn('h-1.5 w-1.5 rounded-full', VARIANT_DOT[variant])} />
      {label}
    </span>
  )
}

/** Map backend status string to a StatusBadge variant + label. */
export function thesisStatusProps(status: string): { variant: StatusVariant; label: string } {
  switch (status) {
    case 'submitted': return { variant: 'pending', label: 'Menunggu Review' }
    case 'approved': return { variant: 'approved', label: 'Disetujui' }
    case 'rejected': return { variant: 'rejected', label: 'Ditolak' }
    case 'cancelled': return { variant: 'rejected', label: 'Dibatalkan' }
    case 'in_progress': return { variant: 'in_progress', label: 'Dalam Bimbingan' }
    case 'seminar_ready': return { variant: 'in_progress', label: 'Siap Seminar' }
    case 'seminar_done': return { variant: 'in_progress', label: 'Pasca Seminar' }
    case 'defense_ready': return { variant: 'in_progress', label: 'Siap Sidang' }
    case 'defense_done': return { variant: 'in_progress', label: 'Pasca Sidang' }
    case 'graduated': return { variant: 'completed', label: 'Lulus' }
    default: return { variant: 'draft', label: status.replace(/_/g, ' ') }
  }
}
