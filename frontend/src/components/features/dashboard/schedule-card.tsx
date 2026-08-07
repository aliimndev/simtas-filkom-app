import { Card, CardContent } from '@/components/ui/card'
import { formatDate } from '@/lib/utils/date'
import type { ScheduleInfo } from '@/types/dashboard'

function ScheduleBlock({ label, info }: { label: string; info?: ScheduleInfo | null }) {
  return (
    <div className="rounded-md border border-border bg-muted/40 p-3">
      <dt className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">{label}</dt>
      <dd className="mt-0.5 text-sm font-medium">
        {info ? (
          <>
            {info.scheduled_at ? formatDate(info.scheduled_at) : '—'}
            {info.room ? ` · ${info.room}` : ''}
          </>
        ) : (
          <span className="font-normal text-muted-foreground">Belum dijadwalkan</span>
        )}
      </dd>
    </div>
  )
}

/** Kartu jadwal ujian mendatang (seminar & sidang). */
export function ScheduleCard({
  seminar,
  defense,
}: {
  seminar?: ScheduleInfo | null
  defense?: ScheduleInfo | null
}) {
  return (
    <Card>
      <CardContent className="p-6">
        <h2 className="text-sm font-semibold">Jadwal</h2>
        <dl className="mt-2 space-y-3">
          <ScheduleBlock label="Seminar" info={seminar} />
          <ScheduleBlock label="Sidang" info={defense} />
        </dl>
      </CardContent>
    </Card>
  )
}
