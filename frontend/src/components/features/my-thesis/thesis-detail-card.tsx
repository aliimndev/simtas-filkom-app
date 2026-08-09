import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import type { Thesis } from '@/types/thesis'

function statusVariant(status: string) {
  switch (status) {
    case 'graduated':
      return 'success'
    case 'cancelled':
      return 'danger'
    case 'rejected':
      return 'danger'
    case 'submitted':
      return 'warning'
    default:
      return 'primary'
  }
}

/** Detail skripsi aktif: judul, status, abstrak, bidang, pembimbing, catatan kaprodi. */
export function ThesisDetailCard({ thesis }: { thesis: Thesis }) {
  return (
    <Card>
      <CardHeader className="flex-row items-start justify-between space-y-0">
        <div className="space-y-1.5">
          <CardTitle className="text-xl leading-snug">{thesis.title}</CardTitle>
          <Badge variant={statusVariant(thesis.status)}>{thesis.status.replace(/_/g, ' ')}</Badge>
        </div>
      </CardHeader>
      <CardContent className="space-y-4">
        {thesis.abstract && (
          <div>
            <p className="mb-1 text-sm font-medium text-muted-foreground">Abstrak</p>
            <p className="text-sm leading-relaxed">{thesis.abstract}</p>
          </div>
        )}
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm font-medium text-muted-foreground">Bidang Keahlian</p>
            <p className="text-sm">{thesis.field_of_study ?? '—'}</p>
          </div>
          <div>
            <p className="text-sm font-medium text-muted-foreground">Pembimbing</p>
            <p className="text-sm">
              {thesis.supervisors?.length ? thesis.supervisors.map((s) => s.full_name).join(', ') : 'Belum ditentukan'}
            </p>
          </div>
        </div>
        {thesis.kaprodi_notes && (
          <div className="rounded-lg border border-secondary/30 bg-secondary-50 p-4">
            <p className="text-sm font-medium text-secondary-foreground">Catatan Kaprodi</p>
            <p className="mt-1 text-sm">{thesis.kaprodi_notes}</p>
          </div>
        )}
      </CardContent>
    </Card>
  )
}
