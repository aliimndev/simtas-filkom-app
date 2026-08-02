'use client'

import { useParams, useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, CheckCircle2, XCircle } from 'lucide-react'
import Link from 'next/link'
import { useState } from 'react'
import { Alert } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { thesisApi } from '@/lib/api/thesis-api'
import { getErrorMessage } from '@/lib/utils/error'
import { formatDate } from '@/lib/utils/date'

export default function ReviewThesisPage() {
  const params = useParams<{ id: string }>()
  const router = useRouter()
  const queryClient = useQueryClient()
  const id = params.id
  const [notes, setNotes] = useState('')
  const [error, setError] = useState<string | null>(null)

  const thesis = useQuery({
    queryKey: ['theses', id],
    queryFn: () => thesisApi.get(id),
    enabled: Boolean(id),
  })

  const review = useMutation({
    mutationFn: (decision: 'approved' | 'rejected') => thesisApi.review(id, { decision, notes: notes || undefined }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['theses'] })
      router.push('/theses')
    },
  })

  const t = thesis.data

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <Link href="/theses" className="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4" /> Kembali ke Daftar
        </Link>
        <h1 className="text-2xl font-bold">Review Pengajuan Skripsi</h1>
      </div>

      {thesis.isLoading ? (
        <p className="py-10 text-center text-sm text-muted-foreground">Memuat…</p>
      ) : !t ? (
        <Alert variant="danger">Skripsi tidak ditemukan.</Alert>
      ) : (
        <>
          <Card>
            <CardHeader className="flex-row items-start justify-between space-y-0">
              <div className="space-y-1.5">
                <CardTitle className="text-lg leading-snug">{t.title}</CardTitle>
                <Badge variant="warning">{t.status.replace(/_/g, ' ')}</Badge>
              </div>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div>
                  <p className="text-sm text-muted-foreground">Mahasiswa</p>
                  <p className="font-medium">{t.student?.full_name ?? '—'}</p>
                  <p className="text-xs text-muted-foreground">{t.student?.nim_nidn ?? ''}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Diajukan</p>
                  <p className="font-medium">{formatDate(t.submitted_at)}</p>
                </div>
                <div>
                  <p className="text-sm text-muted-foreground">Bidang Keahlian</p>
                  <p className="font-medium">{t.field_of_study ?? '—'}</p>
                </div>
              </div>
              <div>
                <p className="mb-1 text-sm font-medium text-muted-foreground">Abstrak</p>
                <p className="text-sm leading-relaxed">{t.abstract}</p>
              </div>
            </CardContent>
          </Card>

          <Card>
            <CardHeader>
              <CardTitle className="text-base">Keputusan Review</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              {error && <Alert variant="danger">{error}</Alert>}
              <div>
                <Label htmlFor="notes">Catatan untuk Mahasiswa</Label>
                <Textarea
                  id="notes"
                  rows={4}
                  placeholder="Catatan revisi atau alasan penolakan…"
                  value={notes}
                  onChange={(e) => setNotes(e.target.value)}
                />
              </div>
              <div className="flex flex-wrap gap-3">
                <Button
                  variant="success"
                  onClick={() => {
                    setError(null)
                    review.mutateAsync('approved').catch((e) => setError(getErrorMessage(e)))
                  }}
                  loading={review.isPending}
                >
                  <CheckCircle2 className="h-4 w-4" /> Setujui
                </Button>
                <Button
                  variant="danger"
                  onClick={() => {
                    setError(null)
                    review.mutateAsync('rejected').catch((e) => setError(getErrorMessage(e)))
                  }}
                  loading={review.isPending}
                >
                  <XCircle className="h-4 w-4" /> Tolak
                </Button>
              </div>
            </CardContent>
          </Card>
        </>
      )}
    </div>
  )
}
