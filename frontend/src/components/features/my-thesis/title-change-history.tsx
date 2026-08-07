import { CheckCircle2, Clock, History, XCircle } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Spinner } from '@/components/ui/spinner'
import { formatDate } from '@/lib/utils/date'
import type { TitleChangeRequest, TitleChangeStatus } from '@/types/title-change'

function titleChangeBadge(status: TitleChangeStatus) {
  switch (status) {
    case 'PENDING':
      return <Badge variant="warning"><Clock className="h-3 w-3" /> Menunggu Persetujuan</Badge>
    case 'APPROVED':
      return <Badge variant="success"><CheckCircle2 className="h-3 w-3" /> Disetujui</Badge>
    case 'REJECTED':
      return <Badge variant="danger"><XCircle className="h-3 w-3" /> Ditolak</Badge>
    default:
      return <Badge variant="muted">Dibatalkan</Badge>
  }
}

/** Riwayat pengajuan & review perubahan judul, dengan aksi batalkan per request PENDING. */
export function TitleChangeHistory({
  requests,
  isLoading,
  onCancel,
}: {
  requests: TitleChangeRequest[]
  isLoading: boolean
  onCancel: (request: TitleChangeRequest) => void
}) {
  return (
    <Card>
      <CardHeader className="flex-row items-center justify-between space-y-0">
        <div className="space-y-1">
          <CardTitle className="flex items-center gap-2 text-base">
            <History className="h-4 w-4 text-primary" /> Riwayat Perubahan Judul
          </CardTitle>
          <CardDescription>Pengajuan dan hasil review perubahan judul skripsi Anda</CardDescription>
        </div>
      </CardHeader>
      <CardContent>
        {isLoading ? (
          <Spinner className="py-8" />
        ) : requests.length === 0 ? (
          <p className="py-6 text-center text-sm text-muted-foreground">Belum ada pengajuan perubahan judul.</p>
        ) : (
          <div className="space-y-3">
            {requests.map((r) => (
              <div key={r.id} className="rounded-xl border border-border p-4">
                <div className="flex flex-wrap items-start justify-between gap-2">
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-medium">Judul baru: {r.requested_title}</p>
                      {titleChangeBadge(r.status)}
                    </div>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Diajukan {formatDate(r.created_at)}
                      {r.requested_by?.full_name && ` oleh ${r.requested_by.full_name}`}
                    </p>
                  </div>
                  {r.status === 'PENDING' && (
                    <Button size="sm" variant="ghost" className="text-danger-700 hover:bg-danger/10" onClick={() => onCancel(r)}>
                      <XCircle className="h-3.5 w-3.5" /> Batalkan
                    </Button>
                  )}
                </div>
                <div className="mt-3 grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
                  <div className="rounded-lg bg-muted/50 p-3">
                    <p className="text-xs font-medium text-muted-foreground">Judul Sebelumnya</p>
                    <p className="mt-0.5 leading-snug line-through decoration-muted-foreground/40">{r.previous_title}</p>
                  </div>
                  <div className="rounded-lg bg-primary-50 p-3">
                    <p className="text-xs font-medium text-primary-700">Judul yang Diajukan</p>
                    <p className="mt-0.5 leading-snug">{r.requested_title}</p>
                  </div>
                </div>
                {r.reason && (
                  <p className="mt-3 text-sm">
                    <span className="font-medium text-muted-foreground">Alasan: </span>
                    {r.reason}
                  </p>
                )}
                {(r.reviewed_by?.full_name || r.review_notes) && (
                  <div className="mt-3 border-t border-border pt-3 text-sm">
                    {r.reviewed_by?.full_name && (
                      <p className="text-muted-foreground">
                        <span className="font-medium text-foreground">Direview oleh:</span> {r.reviewed_by.full_name}
                      </p>
                    )}
                    {r.review_notes && (
                      <p className="mt-1 text-muted-foreground">
                        <span className="font-medium text-foreground">Catatan Pembimbing:</span> {r.review_notes}
                      </p>
                    )}
                  </div>
                )}
              </div>
            ))}
          </div>
        )}
      </CardContent>
    </Card>
  )
}
