'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { PencilLine } from 'lucide-react'
import Link from 'next/link'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ListSkeleton } from '@/components/ui/skeleton'
import { thesisApi } from '@/lib/api/thesis-api'
import { titleChangeApi } from '@/lib/api/title-change-api'
import { toast } from '@/components/ui/toaster'
import { useAuthStore } from '@/lib/stores/auth-store'
import { getErrorMessage } from '@/lib/utils/error'
import { ThesisDetailCard } from '@/components/features/my-thesis/thesis-detail-card'
import { TitleChangeHistory } from '@/components/features/my-thesis/title-change-history'
import { TitleChangeFormDialog } from '@/components/features/my-thesis/title-change-form-dialog'
import { CancelRequestDialog } from '@/components/features/my-thesis/cancel-request-dialog'
import { NoThesisState } from '@/components/features/my-thesis/no-thesis-state'
import type { TitleChangeRequest } from '@/types/title-change'

export default function MyThesisPage() {
  const user = useAuthStore((s) => s.user)
  const queryClient = useQueryClient()
  const [submitOpen, setSubmitOpen] = useState(false)
  const [cancelTarget, setCancelTarget] = useState<TitleChangeRequest | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const theses = useQuery({
    queryKey: ['theses', 'mine'],
    queryFn: () => thesisApi.list({ per_page: 1 }),
    enabled: Boolean(user),
  })

  const thesis = theses.data?.data?.[0]

  const requests = useQuery({
    queryKey: ['title-change', thesis?.id],
    queryFn: () => (thesis?.id ? titleChangeApi.list(thesis.id) : Promise.resolve([])),
    enabled: Boolean(thesis?.id),
  })

  const changeRequests = requests.data ?? []
  const hasPending = changeRequests.some((r) => r.status === 'PENDING')

  // Tombol pengajuan hanya muncul untuk pemilik skripsi dengan status
  // approved/in_progress, ada pembimbing aktif, dan belum ada request PENDING.
  const canRequest =
    user?.role === 'mahasiswa' &&
    thesis &&
    (thesis.status === 'approved' || thesis.status === 'in_progress') &&
    (thesis.supervisors?.length ?? 0) > 0 &&
    !hasPending

  const submit = useMutation({
    mutationFn: (data: { requested_title: string; reason?: string }) => {
      if (!thesis?.id) throw new Error('Tidak ada skripsi aktif')
      return titleChangeApi.create(thesis.id, { requested_title: data.requested_title, reason: data.reason || undefined })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['title-change'] })
      setSubmitOpen(false)
      setActionError(null)
      toast({ variant: 'success', title: 'Perubahan judul diajukan', description: 'Menunggu persetujuan Dosen Pembimbing.' })
    },
  })

  const cancelReq = useMutation({
    mutationFn: (id: string) => titleChangeApi.cancel(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['title-change'] })
      setCancelTarget(null)
      toast({ variant: 'success', title: 'Permintaan dibatalkan' })
    },
  })

  const onSubmit = async (data: { requested_title: string; reason?: string }) => {
    setActionError(null)
    try {
      await submit.mutateAsync(data)
    } catch (err) {
      setActionError(getErrorMessage(err, 'Gagal mengajukan perubahan judul.'))
    }
  }

  if (theses.isLoading) return <ListSkeleton count={4} label="Memuat skripsi…" />

  if (!thesis) return <NoThesisState />

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <h1 className="text-2xl font-bold">Skripsi Saya</h1>
        <div className="flex flex-wrap items-center gap-2">
          {canRequest && (
            <Button size="sm" onClick={() => setSubmitOpen(true)}>
              <PencilLine className="h-4 w-4" /> Ajukan Perubahan Judul
            </Button>
          )}
          <Button asChild variant="outline" size="sm">
            <Link href="/thesis/new">Ajukan Judul Baru</Link>
          </Button>
        </div>
      </div>

      <ThesisDetailCard thesis={thesis} />

      <TitleChangeHistory
        requests={changeRequests}
        isLoading={requests.isLoading}
        onCancel={setCancelTarget}
      />

      <TitleChangeFormDialog
        open={submitOpen}
        onClose={() => setSubmitOpen(false)}
        currentTitle={thesis.title}
        actionError={actionError}
        isSubmitting={submit.isPending}
        onSubmit={onSubmit}
      />

      <CancelRequestDialog
        target={cancelTarget}
        onClose={() => setCancelTarget(null)}
        onConfirm={(id) => cancelReq.mutate(id)}
        isPending={cancelReq.isPending}
        isError={cancelReq.isError}
        errorMessage={cancelReq.isError ? getErrorMessage(cancelReq.error) : null}
      />
    </div>
  )
}
