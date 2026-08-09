import { ArrowRight, Inbox, PencilLine } from 'lucide-react'
import { Alert } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { ListSkeleton } from '@/components/ui/skeleton'
import { formatDate } from '@/lib/utils/date'
import type { TitleChangeRequest } from '@/types/title-change'

function ReviewCard({ request, onReview }: { request: TitleChangeRequest; onReview: () => void }) {
  const r = request
  return (
    <Card className="group">
      <CardContent className="p-4 sm:p-5">
        <div className="flex flex-wrap items-start justify-between gap-3">
          <div className="min-w-0 flex-1">
            <div className="flex flex-wrap items-center gap-2">
              <p className="font-medium">{r.requested_by?.full_name ?? 'Mahasiswa'}</p>
              {r.requested_by?.nim_nidn && (
                <Badge variant="muted">{r.requested_by.nim_nidn}</Badge>
              )}
              <Badge variant="warning">Menunggu</Badge>
            </div>
            <p className="mt-1 text-xs text-muted-foreground">Diajukan {formatDate(r.created_at)}</p>

            <div className="mt-3 flex flex-col gap-2 rounded-lg border border-border p-3 text-sm sm:flex-row sm:items-center">
              <div className="min-w-0 flex-1">
                <p className="text-xs font-medium text-muted-foreground">Judul Saat Ini</p>
                <p className="mt-0.5 line-through decoration-muted-foreground/40">{r.previous_title}</p>
              </div>
              <ArrowRight className="hidden h-4 w-4 shrink-0 text-muted-foreground sm:block" />
              <div className="min-w-0 flex-1">
                <p className="text-xs font-medium text-primary-700">Judul Baru</p>
                <p className="mt-0.5">{r.requested_title}</p>
              </div>
            </div>

            {r.reason && (
              <p className="mt-2 text-sm text-muted-foreground">
                <span className="font-medium text-foreground">Alasan: </span>
                {r.reason}
              </p>
            )}
          </div>
          <Button size="sm" variant="outline" onClick={onReview}>
            <PencilLine className="h-3.5 w-3.5" /> Review
          </Button>
        </div>
      </CardContent>
    </Card>
  )
}

/** Antrian permintaan perubahan judul yang menunggu keputusan pembimbing. */
export function ReviewQueue({
  requests,
  isLoading,
  isError,
  errorMessage,
  onOpenReview,
}: {
  requests: TitleChangeRequest[]
  isLoading: boolean
  isError: boolean
  errorMessage: string | null
  onOpenReview: (request: TitleChangeRequest) => void
}) {
  if (isError) return <Alert variant="danger">{errorMessage}</Alert>

  if (isLoading) return <ListSkeleton count={3} label="Memuat permintaan…" />

  if (requests.length === 0) {
    return (
      <Card>
        <CardContent className="py-14 text-center">
          <Inbox className="mx-auto h-10 w-10 text-muted-foreground" />
          <p className="mt-4 font-semibold">Tidak ada permintaan menunggu</p>
          <p className="mt-1 text-sm text-muted-foreground">
            Belum ada pengajuan perubahan judul dari mahasiswa bimbingan Anda.
          </p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-3">
      {requests.map((r) => (
        <ReviewCard key={r.id} request={r} onReview={() => onOpenReview(r)} />
      ))}
    </div>
  )
}
