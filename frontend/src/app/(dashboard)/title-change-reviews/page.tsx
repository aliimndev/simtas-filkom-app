'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useState } from 'react'
import { RequireAuth } from '@/components/features/require-auth'
import { ReviewQueue } from '@/components/features/title-change-review/review-queue'
import { ReviewDialog, type ReviewDecision } from '@/components/features/title-change-review/review-dialog'
import { titleChangeApi } from '@/lib/api/title-change-api'
import { toast } from '@/components/ui/toaster'
import { getErrorMessage } from '@/lib/utils/error'
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

  const pending = useQuery({
    queryKey: ['title-change', 'pending'],
    queryFn: titleChangeApi.listPending,
  })

  const list = pending.data ?? []

  const review = useMutation({
    mutationFn: ({ id, decision, notes }: { id: string; decision: ReviewDecision; notes?: string }) =>
      decision === 'approve'
        ? titleChangeApi.approve(id, notes ? { review_notes: notes } : undefined)
        : titleChangeApi.reject(id, { review_notes: notes }),
    onSuccess: (_result, vars) => {
      queryClient.invalidateQueries({ queryKey: ['title-change'] })
      queryClient.invalidateQueries({ queryKey: ['theses'] })
      queryClient.invalidateQueries({ queryKey: ['dashboard', 'supervisor'] })
      setReviewTarget(null)
      toast({
        variant: 'success',
        title: vars.decision === 'approve' ? 'Perubahan judul disetujui' : 'Perubahan judul ditolak',
        description:
          vars.decision === 'approve'
            ? 'Judul skripsi telah diperbarui dan mahasiswa akan menerima notifikasi email.'
            : 'Mahasiswa akan menerima notifikasi email beserta catatan Anda.',
      })
    },
  })

  function runReview(decision: ReviewDecision, notes: string) {
    if (!reviewTarget) return
    review.mutate({ id: reviewTarget.id, decision, notes: notes || undefined })
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Review Perubahan Judul</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Permintaan perubahan judul dari mahasiswa bimbingan yang menunggu keputusan Anda
        </p>
      </div>

      <ReviewQueue
        requests={list}
        isLoading={pending.isLoading}
        isError={pending.isError}
        errorMessage={pending.isError ? getErrorMessage(pending.error, 'Gagal memuat antrian review.') : null}
        onOpenReview={setReviewTarget}
      />

      <ReviewDialog
        target={reviewTarget}
        onClose={() => setReviewTarget(null)}
        isPending={review.isPending}
        pendingDecision={review.variables?.decision ?? null}
        isError={review.isError}
        errorMessage={review.isError ? getErrorMessage(review.error) : null}
        onReview={runReview}
      />
    </div>
  )
}
