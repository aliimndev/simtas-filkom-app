import { CheckCircle2, Clock, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { formatDate } from '@/lib/utils/date'
import type { Consultation } from '@/types/consultation'

function statusBadge(status: Consultation['status']) {
  switch (status) {
    case 'approved':
      return <Badge variant="success"><CheckCircle2 className="h-3 w-3" /> Disetujui</Badge>
    case 'rejected':
      return <Badge variant="danger"><XCircle className="h-3 w-3" /> Ditolak</Badge>
    default:
      return <Badge variant="warning"><Clock className="h-3 w-3" /> Menunggu</Badge>
  }
}

/** Satu kartu catatan bimbingan. */
export function ConsultationRow({
  consultation,
  isStudent,
  onApprove,
}: {
  consultation: Consultation
  isStudent: boolean
  onApprove: () => void
}) {
  const c = consultation
  return (
    <Card>
      <CardContent className="p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <p className="font-medium">{c.topic}</p>
              {statusBadge(c.status)}
            </div>
            <p className="mt-1 text-xs text-muted-foreground">
              {formatDate(c.date)}
              {c.student_name && ` · ${c.student_name}`}
              {c.supervisor_name && ` · ${c.supervisor_name}`}
            </p>
          </div>
          {!isStudent && c.status === 'pending' && (
            <Button size="sm" variant="success" onClick={onApprove}>
              <CheckCircle2 className="h-3.5 w-3.5" /> Setujui
            </Button>
          )}
        </div>
        {c.notes && <p className="mt-3 text-sm leading-relaxed">{c.notes}</p>}
        {c.follow_up && (
          <p className="mt-2 text-sm text-muted-foreground">
            <span className="font-medium">Tindak lanjut:</span> {c.follow_up}
          </p>
        )}
      </CardContent>
    </Card>
  )
}
