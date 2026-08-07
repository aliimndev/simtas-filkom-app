import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { cn } from '@/lib/utils/cn'

const STAGE_ORDER = ['submitted', 'approved', 'in_progress', 'seminar_ready', 'seminar_done', 'defense_ready', 'defense_done', 'graduated']

const STAGE_LABELS: Record<string, string> = {
  submitted: 'Pengajuan',
  approved: 'Disetujui',
  in_progress: 'Bimbingan',
  seminar_ready: 'Seminar',
  seminar_done: 'Pasca Seminar',
  defense_ready: 'Sidang',
  defense_done: 'Pasca Sidang',
  graduated: 'Lulus',
}

/** Kartu status skripsi + stepper 8 tahap (scroll horizontal di mobile). */
export function StageStepper({ status, progressPercentage }: { status: string; progressPercentage: number }) {
  const stageIndex = STAGE_ORDER.indexOf(status)
  const stageLabel = STAGE_LABELS[status] ?? status.replace(/_/g, ' ')
  const statusVariant = status === 'cancelled' ? 'danger' : status === 'graduated' ? 'success' : 'primary'

  return (
    <Card>
      <CardContent className="p-6">
        <div className="flex flex-wrap items-center justify-between gap-3">
          <div>
            <h2 className="text-sm font-semibold">Status skripsi</h2>
            <p className="mt-0.5 text-sm text-muted-foreground">
              Tahap {stageLabel} · progres {progressPercentage}%
            </p>
          </div>
          <Badge variant={statusVariant}>{stageLabel}</Badge>
        </div>
        {/* Mobile: horizontal snap scroll | Desktop: flex row */}
        <div className="-mx-1 mt-6 overflow-x-auto pb-1 sm:mx-0 sm:overflow-visible sm:pb-0">
          <div className="flex min-w-160 items-center sm:min-w-0">
            {STAGE_ORDER.map((stage, i) => {
              const done = i <= stageIndex
              return (
                <div key={stage} className="flex flex-1 items-center last:flex-none">
                  <div className="flex flex-col items-center gap-1.5">
                    <div
                      className={cn(
                        'flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold',
                        done ? 'bg-primary text-primary-foreground' : 'bg-surface-hi text-muted-foreground',
                      )}
                    >
                      {i + 1}
                    </div>
                    <span className="whitespace-nowrap text-[11px] text-muted-foreground">{STAGE_LABELS[stage]}</span>
                  </div>
                  {i < STAGE_ORDER.length - 1 && (
                    <div className={cn('mx-2 h-0.5 flex-1 rounded-full', done ? 'bg-primary' : 'bg-border')} />
                  )}
                </div>
              )
            })}
          </div>
        </div>
      </CardContent>
    </Card>
  )
}
