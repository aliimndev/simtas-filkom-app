import { CheckCircle2, Download, Eye, FileText, Loader2, RefreshCcw } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { documentTypeLabel, DOCUMENT_STATUS_LABELS } from '@/types/document'
import { formatDate } from '@/lib/utils/date'
import type { ThesisDocument } from '@/types/document'

function statusBadge(status: ThesisDocument['status']) {
  switch (status) {
    case 'approved':
      return <Badge variant="success"><CheckCircle2 className="h-3 w-3" /> {DOCUMENT_STATUS_LABELS.approved}</Badge>
    case 'revision_required':
      return <Badge variant="danger"><RefreshCcw className="h-3 w-3" /> {DOCUMENT_STATUS_LABELS.revision_required}</Badge>
    default:
      return <Badge variant="warning">{DOCUMENT_STATUS_LABELS.pending_review}</Badge>
  }
}

/** Satu baris kartu dokumen: meta + aksi unduh/review sesuai peran. */
export function DocumentRow({
  doc,
  isStudent,
  downloading,
  onDownload,
  onApprove,
  onRequestRevision,
}: {
  doc: ThesisDocument
  isStudent: boolean
  downloading: boolean
  onDownload: () => void
  onApprove: () => void
  onRequestRevision: () => void
}) {
  return (
    <Card>
      <CardContent className="flex flex-wrap items-center justify-between gap-3 p-4">
        <div className="flex min-w-0 items-center gap-3">
          <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary">
            <FileText className="h-5 w-5" />
          </div>
          <div className="min-w-0">
            <div className="flex items-center gap-2">
              <p className="truncate font-medium">{doc.file_name}</p>
              {statusBadge(doc.status)}
            </div>
            <p className="mt-0.5 text-xs text-muted-foreground">
              {documentTypeLabel(doc.document_type)}
              {doc.chapter_number ? ` Bab ${doc.chapter_number}` : ''} · v{doc.version}
              {doc.file_size != null && ` · ${(doc.file_size / 1024 / 1024).toFixed(2)} MB`} · {doc.created_at ? formatDate(doc.created_at) : ''}
            </p>
            {doc.reviewer_notes && doc.status === 'revision_required' && (
              <p className="mt-1 text-xs text-danger-700">Catatan revisi: {doc.reviewer_notes}</p>
            )}
          </div>
        </div>
        <div className="flex items-center gap-2">
          <Button
            size="sm"
            variant="outline"
            disabled={downloading}
            onClick={onDownload}
          >
            {downloading ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Download className="h-3.5 w-3.5" />}
            Unduh
          </Button>
          {!isStudent && doc.status === 'pending_review' && (
            <>
              <Button size="sm" variant="success" onClick={onApprove}>
                <CheckCircle2 className="h-3.5 w-3.5" /> Setujui
              </Button>
              <Button size="sm" variant="ghost" onClick={onRequestRevision}>
                <RefreshCcw className="h-3.5 w-3.5" /> Minta Revisi
              </Button>
            </>
          )}
          {!isStudent && doc.status !== 'pending_review' && (
            <Button size="sm" variant="ghost" disabled={downloading} onClick={onDownload}>
              <Eye className="h-3.5 w-3.5" /> Lihat
            </Button>
          )}
        </div>
      </CardContent>
    </Card>
  )
}
