'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowRight, CheckCircle2, Inbox, PencilLine, XCircle } from 'lucide-react'
import { useState } from 'react'
import { Alert } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Dialog } from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { RequireAuth } from '@/components/features/require-auth'
import { ListSkeleton } from '@/components/ui/skeleton'
import { Textarea } from '@/components/ui/textarea'
import { titleChangeApi } from '@/lib/api/title-change-api'
import { toast } from '@/components/ui/toaster'
import { getErrorMessage } from '@/lib/utils/error'
import { formatDate } from '@/lib/utils/date'
import type { TitleChangeRequest } from '@/types/title-change'

export default function TitleChangeReviewsPage() {
  return (
    <RequireAuth roles={['dosen_pembimbing']}>
      <TitleChangeReviewsInner />
    </RequireAuth>
  )
}

function TitleChangeReviewsInner() {
  const queryClient = useQueryClient()
  const [reviewTarget, setReviewTarget] = useState<TitleChangeRequest | null>(null)
  const [notes, setNotes] = useState('')
  const [notesTouched, setNotesTouched] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  const pending = useQuery({
    queryKey: ['title-change', 'pending'],
    queryFn: titleChangeApi.listPending,
  })

  const list = pending.data ?? []

  const review = useMutation({
    mutationFn: ({ id, decision }: { id: string; decision: 'approve' | 'reject' }) =>
      decision === 'approve'
        ? titleChangeApi.approve(id, notes ? { review_notes: notes } : undefined)
        : titleChangeApi.reject(id, { review_notes: notes }),
    onSuccess: (_result, vars) => {
      queryClient.invalidateQueries({ queryKey: ['title-change'] })
      queryClient.invalidateQueries({ queryKey: ['theses'] })
      queryClient.invalidateQueries({ queryKey: ['dashboard', 'supervisor'] })
      setReviewTarget(null)
      setNotes('')
      setNotesTouched(false)
      setActionError(null)
      toast({
        variant: 'success',
        title: vars.decision === 'approve' ? 'Perubahan judul disetujui' : 'Perubahan judul ditolak',
        description:
          vars.decision === 'approve'
            ? 'Judul skripsi telah diperbarui dan mahasiswa akan menerima notifikasi email.'
            : 'Mahasiswa akan menerima notifikasi email beserta catatan Anda.',
      })
    },
    onError: (e) => setActionError(getErrorMessage(e)),
  })

  function openReview(r: TitleChangeRequest) {
    setReviewTarget(r)
    setNotes('')
    setNotesTouched(false)
    setActionError(null)
  }

  function run(decision: 'approve' | 'reject') {
    if (!reviewTarget) return
    if (decision === 'reject' && !notes.trim()) {
      setNotesTouched(true)
      return
    }
    setActionError(null)
    review.mutate({ id: reviewTarget.id, decision })
  }

  const notesInvalid = notesTouched && !notes.trim()

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Review Perubahan Judul</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Permintaan perubahan judul dari mahasiswa bimbingan yang menunggu keputusan Anda
        </p>
      </div>

      {pending.isError ? (
        <Alert variant="danger">{getErrorMessage(pending.error, 'Gagal memuat antrian review.')}</Alert>
      ) : pending.isLoading ? (
        <ListSkeleton count={3} label="Memuat permintaan…" />
      ) : list.length === 0 ? (
        <Card>
          <CardContent className="py-14 text-center">
            <Inbox className="mx-auto h-10 w-10 text-muted-foreground" />
            <p className="mt-4 font-semibold">Tidak ada permintaan menunggu</p>
            <p className="mt-1 text-sm text-muted-foreground">
              Belum ada pengajuan perubahan judul dari mahasiswa bimbingan Anda. 🎉
            </p>
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-3">
          {list.map((r) => (
            <Card key={r.id} className="group">
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
                  <Button size="sm" variant="outline" onClick={() => openReview(r)}>
                    <PencilLine className="h-3.5 w-3.5" /> Review
                  </Button>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* ── Dialog: Konfirmasi Review ──────────────────────────────────── */}
      <Dialog
        open={Boolean(reviewTarget)}
        onClose={() => setReviewTarget(null)}
        title="Konfirmasi Perubahan Judul"
        description="Tinjau pengajuan mahasiswa sebelum mengambil keputusan."
      >
        {reviewTarget && (
          <div className="space-y-4">
            <div className="grid grid-cols-1 gap-3 text-sm sm:grid-cols-2">
              <div className="rounded-lg bg-muted/50 p-3">
                <p className="text-xs font-medium text-muted-foreground">Judul Sebelumnya</p>
                <p className="mt-0.5 leading-snug">{reviewTarget.previous_title}</p>
              </div>
              <div className="rounded-lg bg-primary-50 p-3">
                <p className="text-xs font-medium text-primary-700">Judul Baru</p>
                <p className="mt-0.5 leading-snug">{reviewTarget.requested_title}</p>
              </div>
            </div>
            {reviewTarget.reason && (
              <div className="rounded-lg border border-border p-3 text-sm">
                <p className="text-xs font-medium text-muted-foreground">Alasan Mahasiswa</p>
                <p className="mt-0.5">{reviewTarget.reason}</p>
              </div>
            )}
            <div>
              <Label htmlFor="review-notes">Catatan Pembimbing</Label>
              <Textarea
                id="review-notes"
                rows={3}
                placeholder="Catatan untuk mahasiswa (wajib jika menolak)"
                value={notes}
                invalid={notesInvalid}
                onChange={(e) => {
                  setNotes(e.target.value)
                  if (notesTouched) setNotesTouched(false)
                }}
              />
              {notesInvalid && <p className="mt-1 text-xs text-danger">Catatan wajib diisi saat menolak perubahan judul.</p>}
            </div>
            {actionError && <Alert variant="danger">{actionError}</Alert>}
            <div className="flex flex-wrap items-center justify-between gap-2 border-t border-border pt-4">
              <Button type="button" variant="ghost" onClick={() => setReviewTarget(null)}>
                Kembali
              </Button>
              <div className="flex gap-2">
                <Button
                  type="button"
                  variant="danger"
                  loading={review.isPending && review.variables?.decision === 'reject'}
                  disabled={review.isPending}
                  onClick={() => run('reject')}
                >
                  <XCircle className="h-4 w-4" /> Tolak
                </Button>
                <Button
                  type="button"
                  variant="success"
                  loading={review.isPending && review.variables?.decision === 'approve'}
                  disabled={review.isPending}
                  onClick={() => run('approve')}
                >
                  <CheckCircle2 className="h-4 w-4" /> Setujui
                </Button>
              </div>
            </div>
          </div>
        )}
      </Dialog>
    </div>
  )
}
