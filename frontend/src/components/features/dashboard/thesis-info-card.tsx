import { Card, CardContent } from '@/components/ui/card'
import { formatDate } from '@/lib/utils/date'
import type { StudentDashboard } from '@/types/dashboard'

/** Ringkasan skripsi: judul, pembimbing, dan bimbingan terakhir. */
export function ThesisInfoCard({ thesis }: { thesis: StudentDashboard }) {
  return (
    <Card className="lg:col-span-2">
      <CardContent className="p-6">
        <h2 className="text-sm font-semibold">Skripsi Saya</h2>
        <dl className="mt-2 divide-y divide-border">
          <div className="flex flex-col gap-1 py-3 sm:flex-row sm:items-baseline sm:justify-between sm:gap-6">
            <dt className="shrink-0 text-sm text-muted-foreground">Judul</dt>
            <dd className="text-sm font-medium text-foreground">{thesis.title}</dd>
          </div>
          <div className="flex flex-col gap-1 py-3 sm:flex-row sm:items-baseline sm:justify-between sm:gap-6">
            <dt className="shrink-0 text-sm text-muted-foreground">Pembimbing</dt>
            <dd className="text-sm font-medium text-foreground">
              {thesis.supervisors?.length ? thesis.supervisors.map((s) => s.full_name).join(', ') : 'Belum ditentukan'}
            </dd>
          </div>
          <div className="flex flex-col gap-1 py-3 sm:flex-row sm:items-baseline sm:justify-between sm:gap-6">
            <dt className="shrink-0 text-sm text-muted-foreground">Bimbingan terakhir</dt>
            <dd className="text-sm font-medium text-foreground">
              {thesis.last_consultation ? formatDate(thesis.last_consultation) : 'Belum ada'}
            </dd>
          </div>
        </dl>
      </CardContent>
    </Card>
  )
}
