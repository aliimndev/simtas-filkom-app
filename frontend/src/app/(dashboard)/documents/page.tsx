'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { FileText } from 'lucide-react'
import { useRef, useState } from 'react'
import { Card, CardContent } from '@/components/ui/card'
import { Spinner } from '@/components/ui/spinner'
import { ListSkeleton } from '@/components/ui/skeleton'
import { documentApi } from '@/lib/api/document-api'
import { getErrorMessage } from '@/lib/utils/error'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useThesisPicker } from '@/lib/hooks/use-thesis-picker'
import { DocumentUploadControls } from '@/components/features/documents-page/document-upload-controls'
import { DocumentFilter } from '@/components/features/documents-page/document-filter'
import { DocumentRow } from '@/components/features/documents-page/document-row'
import type { DocumentType } from '@/types/document'

export default function DocumentsPage() {
  const user = useAuthStore((s) => s.user)
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

  if (thesesLoading) return <ListSkeleton count={4} label="Memuat dokumen…" />

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
          <p className="landing-eyebrow">Dokumen Skripsi</p>
          <h1 className="mt-2 text-balance landing-heading text-2xl">
            {isStudent ? 'Unggah dan Pantau <span className="accent-text italic">Dokumen</span>' : 'Review Dokumen Mahasiswa'}
          </h1>
          <p className="mt-1.5 text-sm text-st-muted">
            {isStudent ? 'Unggah dan pantau dokumen skripsi Anda' : 'Review dokumen yang diunggah mahasiswa bimbingan'}
          </p>
          {!isStudent && picker}
        </div>
        {isStudent && (
          <DocumentUploadControls
            uploadType={uploadType}
            onUploadTypeChange={setUploadType}
            chapterNumber={chapterNumber}
            onChapterNumberChange={setChapterNumber}
            uploading={uploading}
            onUploadFile={handleUpload}
            onCancelUpload={cancelUpload}
          />
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

      <DocumentFilter selectedType={selectedType} onSelect={setSelectedType} />

      <div className="space-y-3">
        {docs.isLoading && <Spinner />}
        {list.map((doc) => (
          <DocumentRow
            key={doc.id}
            doc={doc}
            isStudent={isStudent}
            downloading={downloadingId === doc.id}
            onDownload={() => handleDownload(myThesis.id, doc.id)}
            onApprove={() => review.mutate({ id: doc.id, decision: 'approved' })}
            onRequestRevision={() =>
              review.mutate({ id: doc.id, decision: 'revision_required', notes: 'Silakan perbaiki sesuai catatan.' })
            }
          />
        ))}
        {!docs.isLoading && list.length === 0 && (
          <div className="py-12 text-center">
            <FileText className="mx-auto h-10 w-10 text-st-muted/40" />
            <p className="mt-3 landing-heading text-lg">Belum ada <span className="accent-text italic">dokumen</span></p>
            <p className="mt-1 text-sm text-st-muted">Unggah dokumen skripsi pertama Anda untuk memulai.</p>
          </div>
        )}
      </div>
    </div>
  )
}
