'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Download, FileText, Upload, CheckCircle2, RefreshCcw, Eye, Loader2, X } from 'lucide-react'
import { useRef, useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { documentApi } from '@/lib/api/document-api'
import { documentTypeLabel, DOCUMENT_STATUS_LABELS, type DocumentType } from '@/types/document'
import { formatDate } from '@/lib/utils/date'
import { getErrorMessage } from '@/lib/utils/error'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useThesisPicker } from '@/lib/hooks/use-thesis-picker'

// Document types match backend constants (Job 07/21).
const TYPES: { value: DocumentType; label: string }[] = [
  { value: 'proposal', label: 'Proposal' },
  { value: 'draft_chapter', label: 'Draft Bab' },
  { value: 'seminar_doc', label: 'Dokumen Seminar' },
  { value: 'defense_doc', label: 'Dokumen Sidang' },
  { value: 'final_thesis', label: 'Skripsi Final' },
  { value: 'revision_sheet', label: 'Lembar Revisi' },
  { value: 'endorsement_letter', label: 'Surat Pengesahan' },
]

function statusBadge(status: string) {
  switch (status) {
    case 'approved':
      return <Badge variant="success"><CheckCircle2 className="h-3 w-3" /> {DOCUMENT_STATUS_LABELS.approved}</Badge>
    case 'revision_required':
      return <Badge variant="danger"><RefreshCcw className="h-3 w-3" /> {DOCUMENT_STATUS_LABELS.revision_required}</Badge>
    default:
      return <Badge variant="warning">{DOCUMENT_STATUS_LABELS.pending_review}</Badge>
  }
}

export default function DocumentsPage() {
  const { user } = useAuthStore()
  const queryClient = useQueryClient()
  const isStudent = user?.role === 'mahasiswa'
  const [selectedType, setSelectedType] = useState('')
  const [uploadType, setUploadType] = useState<DocumentType>('proposal')
  const [chapterNumber, setChapterNumber] = useState('1')
  const [uploading, setUploading] = useState(false)
  const [uploadProgress, setUploadProgress] = useState(0)
  const uploadAbortRef = useRef<AbortController | null>(null)
  const [error, setError] = useState<string | null>(null)
  const [downloadingId, setDownloadingId] = useState<string | null>(null)

  const { thesis: myThesis, picker, isLoading: thesesLoading } = useThesisPicker()

  const docs = useQuery({
    queryKey: ['documents', myThesis?.id, selectedType],
    queryFn: () =>
      myThesis?.id
        ? documentApi.list(myThesis.id, { per_page: 50, document_type: selectedType || undefined })
        : Promise.resolve(null),
    enabled: Boolean(myThesis?.id),
  })

  const review = useMutation({
    mutationFn: ({ id, decision, notes }: { id: string; decision: 'approved' | 'revision_required'; notes?: string }) =>
      documentApi.review(id, decision, notes),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['documents'] })
    },
    onError: (e) => setError(getErrorMessage(e)),
  })

  const list = docs.data?.data ?? []

  async function handleUpload(file: File) {
    if (!myThesis?.id) return
    setError(null)
    setUploading(true)
    setUploadProgress(0)
    const controller = new AbortController()
    uploadAbortRef.current = controller
    try {
      await documentApi.upload(
        myThesis.id,
        file,
        uploadType,
        uploadType === 'draft_chapter' ? Number(chapterNumber) : undefined,
        setUploadProgress,
        controller.signal,
      )
      queryClient.invalidateQueries({ queryKey: ['documents'] })
    } catch (e) {
      if (!controller.signal.aborted) {
        setError(getErrorMessage(e))
      }
    } finally {
      uploadAbortRef.current = null
      setUploading(false)
      setUploadProgress(0)
    }
  }

  function cancelUpload() {
    uploadAbortRef.current?.abort()
  }

  async function handleDownload(thesisId: string, id: string) {
    setDownloadingId(id)
    setError(null)
    try {
      await documentApi.download(thesisId, id)
    } catch (e) {
      setError(getErrorMessage(e))
    } finally {
      setDownloadingId(null)
    }
  }

  if (thesesLoading) return <Spinner label="Memuat dokumen…" />

  if (!myThesis) {
    return (
      <Card>
        <CardContent className="py-14 text-center">
          <FileText className="mx-auto h-10 w-10 text-muted-foreground" />
          <p className="mt-4 font-semibold">Belum ada skripsi aktif</p>
          <p className="mt-1 text-sm text-muted-foreground">Ajukan skripsi terlebih dahulu untuk mengunggah dokumen.</p>
        </CardContent>
      </Card>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">Dokumen</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {isStudent ? 'Unggah dan pantau dokumen skripsi Anda' : 'Review dokumen yang diunggah mahasiswa bimbingan'}
          </p>
          {!isStudent && picker}
        </div>
        {isStudent && (
          <div className="flex flex-wrap items-end gap-2">
            <div>
              <Label htmlFor="type">Jenis Dokumen</Label>
              <Select id="type" value={uploadType} onChange={(e) => setUploadType(e.target.value as DocumentType)} className="w-44">
                {TYPES.map((t) => (
                  <option key={t.value} value={t.value}>{t.label}</option>
                ))}
              </Select>
            </div>
            {uploadType === 'draft_chapter' && (
              <div>
                <Label htmlFor="chapter">Bab</Label>
                <Select id="chapter" value={chapterNumber} onChange={(e) => setChapterNumber(e.target.value)} className="w-20">
                  {['1', '2', '3', '4', '5'].map((n) => (
                    <option key={n} value={n}>{n}</option>
                  ))}
                </Select>
              </div>
            )}
            {uploading ? (
              <Button size="sm" variant="ghost" onClick={cancelUpload}>
                <X className="h-4 w-4" /> Batalkan
              </Button>
            ) : (
              <label className="inline-flex h-10 cursor-pointer items-center gap-2 rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary-700">
                <Upload className="h-4 w-4" />
                Unggah
                <input
                  type="file"
                  accept=".pdf"
                  className="hidden"
                  onChange={(e) => {
                    const f = e.target.files?.[0]
                    if (f) handleUpload(f)
                    e.target.value = ''
                  }}
                />
              </label>
            )}
          </div>
        )}
      </div>

      {uploading && (
        <div className="space-y-1">
          <div className="h-2 w-full overflow-hidden rounded-full bg-muted">
            <div
              className="h-full rounded-full bg-primary transition-all duration-300"
              style={{ width: `${uploadProgress}%` }}
            />
          </div>
          <p className="text-xs text-muted-foreground">{uploadProgress}% terunggah</p>
        </div>
      )}

      {error && <p className="text-sm text-danger">{error}</p>}

      <div className="flex flex-wrap gap-2">
        <Button
          size="sm"
          variant={selectedType === '' ? 'primary' : 'outline'}
          onClick={() => setSelectedType('')}
        >
          Semua
        </Button>
        {TYPES.map((t) => (
          <Button
            key={t.value}
            size="sm"
            variant={selectedType === t.value ? 'primary' : 'outline'}
            onClick={() => setSelectedType(t.value)}
          >
            {t.label}
          </Button>
        ))}
      </div>

      <div className="space-y-3">
        {docs.isLoading && <Spinner />}
        {list.map((doc) => (
          <Card key={doc.id}>
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
                    {doc.file_size != null && ` · ${(doc.file_size / 1024 / 1024).toFixed(2)} MB`} · {formatDate(doc.created_at)}
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
                  disabled={downloadingId === doc.id}
                  onClick={() => handleDownload(myThesis.id, doc.id)}
                >
                  {downloadingId === doc.id ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : <Download className="h-3.5 w-3.5" />}
                  Unduh
                </Button>
                {!isStudent && doc.status === 'pending_review' && (
                  <>
                    <Button
                      size="sm"
                      variant="success"
                      onClick={() => review.mutate({ id: doc.id, decision: 'approved' })}
                    >
                      <CheckCircle2 className="h-3.5 w-3.5" /> Setujui
                    </Button>
                    <Button
                      size="sm"
                      variant="ghost"
                      onClick={() => review.mutate({ id: doc.id, decision: 'revision_required', notes: 'Silakan perbaiki sesuai catatan.' })}
                    >
                      <RefreshCcw className="h-3.5 w-3.5" /> Minta Revisi
                    </Button>
                  </>
                )}
                {!isStudent && doc.status !== 'pending_review' && (
                  <Button
                    size="sm"
                    variant="ghost"
                    disabled={downloadingId === doc.id}
                    onClick={() => handleDownload(myThesis.id, doc.id)}
                  >
                    <Eye className="h-3.5 w-3.5" /> Lihat
                  </Button>
                )}
              </div>
            </CardContent>
          </Card>
        ))}
        {!docs.isLoading && list.length === 0 && (
          <p className="py-10 text-center text-sm text-muted-foreground">Belum ada dokumen.</p>
        )}
      </div>
    </div>
  )
}
