'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { BookOpen, CheckCircle2, Clock, History, PencilLine, XCircle } from 'lucide-react'
import Link from 'next/link'
import { useState } from 'react'
import { Alert } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Spinner } from '@/components/ui/spinner'
import { ListSkeleton } from '@/components/ui/skeleton'
import { Textarea } from '@/components/ui/textarea'
import { thesisApi } from '@/lib/api/thesis-api'
import { titleChangeApi } from '@/lib/api/title-change-api'
import { toast } from '@/components/ui/toaster'
import { useAuthStore } from '@/lib/stores/auth-store'
import { getErrorMessage } from '@/lib/utils/error'
import { formatDate } from '@/lib/utils/date'
import type { TitleChangeRequest, TitleChangeStatus } from '@/types/title-change'

function statusVariant(status: string) {
  switch (status) {
    case 'graduated':
      return 'success'
    case 'cancelled':
      return 'danger'
    case 'rejected':
      return 'danger'
    case 'submitted':
      return 'warning'
    default:
      return 'primary'
  }
}

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

const submitSchema = z.object({
  requested_title: z
    .string()
    .min(1, 'Judul baru wajib diisi')
    .max(500, 'Judul maksimal 500 karakter')
    .refine((v) => v.trim().split(/\s+/).filter(Boolean).length >= 10, 'Judul minimal 10 kata'),
  reason: z.string().optional(),
})

type SubmitForm = z.infer<typeof submitSchema>

export default function MyThesisPage() {
  const { user } = useAuthStore()
  const queryClient = useQueryClient()
  const [submitOpen, setSubmitOpen] = useState(false)
  const [cancelTarget, setCancelTarget] = useState<TitleChangeRequest | null>(null)
  const [actionError, setActionError] = useState<string | null>(null)

  const theses = useQuery({
    queryKey: ['theses', 'mine'],
    queryFn: () => thesisApi.list({ per_page: 1 }),
    enabled: Boolean(user),
  })

  const list = theses.data?.data ?? []
  const thesis = list[0]

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

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<SubmitForm>({ resolver: zodResolver(submitSchema) })

  const submit = useMutation({
    mutationFn: (data: SubmitForm) => {
      if (!thesis?.id) throw new Error('Tidak ada skripsi aktif')
      return titleChangeApi.create(thesis.id, { requested_title: data.requested_title, reason: data.reason || undefined })
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['title-change'] })
      setSubmitOpen(false)
      reset()
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

  const onSubmit = handleSubmit(async (data) => {
    setActionError(null)
    try {
      await submit.mutateAsync(data)
    } catch (err) {
      setActionError(getErrorMessage(err, 'Gagal mengajukan perubahan judul.'))
    }
  })

  if (theses.isLoading) return <ListSkeleton count={4} label="Memuat skripsi…" />

  if (!thesis) {
    return (
      <div className="mx-auto max-w-lg space-y-6 py-10 text-center">
        <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-50 text-primary">
          <BookOpen className="h-8 w-8" />
        </div>
        <div>
          <h1 className="text-xl font-bold">Anda belum memiliki skripsi</h1>
          <p className="mt-2 text-sm text-muted-foreground">
            Ajukan judul skripsi Anda untuk memulai perjalanan Tugas Akhir.
          </p>
        </div>
        <Button asChild size="lg">
          <Link href="/thesis/new">Ajukan Judul Skripsi</Link>
        </Button>
      </div>
    )
  }

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

      <Card>
        <CardHeader className="flex-row items-start justify-between space-y-0">
          <div className="space-y-1.5">
            <CardTitle className="text-xl leading-snug">{thesis.title}</CardTitle>
            <Badge variant={statusVariant(thesis.status)}>{thesis.status.replace(/_/g, ' ')}</Badge>
          </div>
        </CardHeader>
        <CardContent className="space-y-4">
          {thesis.abstract && (
            <div>
              <p className="mb-1 text-sm font-medium text-muted-foreground">Abstrak</p>
              <p className="text-sm leading-relaxed">{thesis.abstract}</p>
            </div>
          )}
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <div>
              <p className="text-sm font-medium text-muted-foreground">Bidang Keahlian</p>
              <p className="text-sm">{thesis.field_of_study ?? '—'}</p>
            </div>
            <div>
              <p className="text-sm font-medium text-muted-foreground">Pembimbing</p>
              <p className="text-sm">
                {thesis.supervisors?.length ? thesis.supervisors.map((s) => s.full_name).join(', ') : 'Belum ditentukan'}
              </p>
            </div>
          </div>
          {thesis.kaprodi_notes && (
            <div className="rounded-lg border border-secondary/30 bg-secondary-50 p-4">
              <p className="text-sm font-medium text-secondary-foreground">Catatan Kaprodi</p>
              <p className="mt-1 text-sm">{thesis.kaprodi_notes}</p>
            </div>
          )}
        </CardContent>
      </Card>

      {/* ── Riwayat Perubahan Judul ─────────────────────────────────────── */}
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
          {requests.isLoading ? (
            <Spinner className="py-8" />
          ) : changeRequests.length === 0 ? (
            <p className="py-6 text-center text-sm text-muted-foreground">Belum ada pengajuan perubahan judul.</p>
          ) : (
            <div className="space-y-3">
              {changeRequests.map((r) => (
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
                      <Button size="sm" variant="ghost" className="text-danger-700 hover:bg-danger/10" onClick={() => setCancelTarget(r)}>
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

      {/* ── Dialog: Ajukan Perubahan Judul ─────────────────────────────── */}
      <Dialog
        open={submitOpen}
        onClose={() => setSubmitOpen(false)}
        title="Ajukan Perubahan Judul"
        description="Perubahan judul akan diproses oleh Dosen Pembimbing Anda."
      >
        <form onSubmit={onSubmit} noValidate className="space-y-4">
          {actionError && <Alert variant="danger">{actionError}</Alert>}
          <div>
            <Label>Judul Saat Ini</Label>
            <Input value={thesis.title} readOnly className="bg-muted/60 text-muted-foreground" />
          </div>
          <div>
            <Label htmlFor="requested_title" required>Judul Baru</Label>
            <Textarea
              id="requested_title"
              rows={3}
              placeholder="Tulis judul baru skripsi Anda (minimal 10 kata)"
              invalid={!!errors.requested_title}
              {...register('requested_title')}
            />
            {errors.requested_title && <p className="mt-1 text-xs text-danger">{errors.requested_title.message}</p>}
          </div>
          <div>
            <Label htmlFor="reason">Alasan Perubahan</Label>
            <Textarea
              id="reason"
              rows={2}
              placeholder="Alasan mengajukan perubahan judul (opsional)"
              {...register('reason')}
            />
          </div>
          <div className="flex flex-wrap justify-end gap-2 pt-1">
            <Button type="button" variant="ghost" onClick={() => setSubmitOpen(false)}>Batal</Button>
            <Button type="submit" loading={submit.isPending}>
              <PencilLine className="h-4 w-4" /> Ajukan
            </Button>
          </div>
        </form>
      </Dialog>

      {/* ── Dialog: Batalkan Permintaan ────────────────────────────────── */}
      <Dialog
        open={Boolean(cancelTarget)}
        onClose={() => setCancelTarget(null)}
        title="Batalkan Permintaan Perubahan Judul?"
        description="Permintaan yang sedang diproses akan ditarik dan tidak dapat dikembalikan."
      >
        {cancelTarget && (
          <div className="space-y-4">
            <div className="rounded-lg bg-muted/50 p-3 text-sm">
              <p className="text-xs font-medium text-muted-foreground">Judul yang diajukan</p>
              <p className="mt-0.5">{cancelTarget.requested_title}</p>
            </div>
            {cancelReq.isError && <Alert variant="danger">{getErrorMessage(cancelReq.error)}</Alert>}
            <div className="flex flex-wrap justify-end gap-2">
              <Button type="button" variant="ghost" onClick={() => setCancelTarget(null)}>Kembali</Button>
              <Button
                type="button"
                variant="danger"
                loading={cancelReq.isPending}
                onClick={() => cancelReq.mutate(cancelTarget.id)}
              >
                Ya, Batalkan
              </Button>
            </div>
          </div>
        )}
      </Dialog>
    </div>
  )
}
