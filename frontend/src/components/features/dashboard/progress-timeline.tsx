import { Check } from 'lucide-react'
import { cn } from '@/lib/utils/cn'

const STAGES = [
  { key: 'submitted', label: 'Pengajuan' },
  { key: 'approved', label: 'Persetujuan' },
  { key: 'in_progress', label: 'Bimbingan' },
  { key: 'seminar_ready', label: 'Seminar' },
  { key: 'seminar_done', label: 'Pasca Seminar' },
  { key: 'defense_ready', label: 'Sidang' },
  { key: 'defense_done', label: 'Pasca Sidang' },
  { key: 'graduated', label: 'Lulus' },
] as const

type StageState = 'completed' | 'current' | 'upcoming' | 'blocked'

function resolveState(stageIndex: number, currentIndex: number, status: string): StageState {
  if (status === 'cancelled' || status === 'rejected') {
    if (stageIndex < currentIndex) return 'completed'
    if (stageIndex === currentIndex) return 'blocked'
    return 'upcoming'
  }
  if (stageIndex < currentIndex) return 'completed'
  if (stageIndex === currentIndex) return 'current'
  return 'upcoming'
}

export function ProgressTimeline({ status }: { status: string }) {
  const currentIndex = STAGES.findIndex((s) => s.key === status)
  const resolvedIndex = currentIndex === -1 ? 0 : currentIndex

  return (
    <div className="w-full">
      <h2 className="mb-4 landing-heading text-lg">Progres <span className="accent-text italic">Tugas Akhir</span></h2>
      {/* Desktop: horizontal */}
      <div className="hidden sm:block">
        <div className="flex items-start">
          {STAGES.map((stage, i) => {
            const state = resolveState(i, resolvedIndex, status)
            return (
              <div key={stage.key} className="flex flex-1 items-start last:flex-none">
                <div className="flex flex-col items-center gap-2">
                  <StageNode state={state} />
                  <span
                    className={cn(
                      'max-w-16 text-center text-[11px] leading-tight',
                      state === 'current' ? 'font-medium text-(--st-accent-to)' :
                      state === 'completed' ? 'text-st-text' :
                      state === 'blocked' ? 'text-danger' :
                      'text-st-muted',
                    )}
                  >
                    {stage.label}
                  </span>
                </div>
                {i < STAGES.length - 1 && (
                  <div
                    className={cn(
                      'mt-3 h-px flex-1',
                      state === 'completed' ? 'accent-gradient' :
                      state === 'blocked' ? 'bg-danger/40' :
                      'bg-st-stroke',
                    )}
                  />
                )}
              </div>
            )
          })}
        </div>
      </div>
      {/* Mobile: vertical */}
      <div className="sm:hidden">
        <div className="flex flex-col gap-0">
          {STAGES.map((stage, i) => {
            const state = resolveState(i, resolvedIndex, status)
            return (
              <div key={stage.key} className="flex items-start gap-3">
                <div className="flex flex-col items-center">
                  <StageNode state={state} />
                  {i < STAGES.length - 1 && (
                    <div
                      className={cn(
                        'w-px flex-1 min-h-6',
                        state === 'completed' ? 'accent-gradient' :
                        state === 'blocked' ? 'bg-danger/40' :
                        'bg-st-stroke',
                      )}
                    />
                  )}
                </div>
                <span
                  className={cn(
                    'pb-5 pt-0.5 text-sm',
                    state === 'current' ? 'font-medium text-(--st-accent-to)' :
                    state === 'completed' ? 'text-st-text' :
                    state === 'blocked' ? 'text-danger' :
                    'text-st-muted',
                  )}
                >
                  {stage.label}
                </span>
              </div>
            )
          })}
        </div>
      </div>
    </div>
  )
}

function StageNode({ state }: { state: StageState }) {
  if (state === 'completed') {
    return (
      <div className="accent-gradient flex h-6 w-6 shrink-0 items-center justify-center rounded-full text-st-bg">
        <Check className="h-3.5 w-3.5" strokeWidth={2.5} />
      </div>
    )
  }
  if (state === 'current') {
    return (
      <div className="relative flex h-6 w-6 shrink-0 items-center justify-center">
        <div className="absolute inset-0 animate-ping rounded-full bg-primary/20" />
        <div className="h-3 w-3 rounded-full accent-gradient" />
      </div>
    )
  }
  if (state === 'blocked') {
    return (
      <div className="flex h-6 w-6 shrink-0 items-center justify-center rounded-full border-2 border-danger bg-danger-50">
        <div className="h-2 w-2 rounded-full bg-danger" />
      </div>
    )
  }
  return (
    <div className="flex h-6 w-6 shrink-0 items-center justify-center">
      <div className="h-2.5 w-2.5 rounded-full bg-st-stroke" />
    </div>
  )
}
