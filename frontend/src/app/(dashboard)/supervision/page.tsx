'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { MessageSquarePlus, CheckCircle2, XCircle, Clock } from 'lucide-react'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Alert } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { Spinner } from '@/components/ui/spinner'
import { ListSkeleton } from '@/components/ui/skeleton'
import { consultationApi } from '@/lib/api/consultation-api'
import { getErrorMessage } from '@/lib/utils/error'
import { formatDate } from '@/lib/utils/date'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useThesisPicker } from '@/lib/hooks/use-thesis-picker'
import type { Consultation } from '@/types/consultation'

const consultSchema = z.object({
  date: z.string().min(1, 'Tanggal wajib diisi'),
  topic: z.string().min(3, 'Topik wajib diisi'),
  notes: z.string().optional(),
  follow_up: z.string().optional(),
})

type ConsultForm = z.infer<typeof consultSchema>

function statusBadge(status: string) {
  switch (status) {
    case 'approved':
      return <Badge variant="success"><CheckCircle2 className="h-3 w-3" /> Disetujui</Badge>
    case 'rejected':
      return <Badge variant="danger"><XCircle className="h-3 w-3" /> Ditolak</Badge>
    default:
      return <Badge variant="warning"><Clock className="h-3 w-3" /> Menunggu</Badge>
  }
}

export default function SupervisionPage() {
  const { user } = useAuthStore()
  const queryClient = useQueryClient()
  const isStudent = user?.role === 'mahasiswa'
  const [showForm, setShowForm] = useState(false)
  const [actionError, setActionError] = useState<string | null>(null)

  // Mahasiswa: skripsi sendiri. Dosen/staf: pilih mahasiswa bimbingan lewat picker.
  const { thesis: myThesis, picker, isLoading: thesesLoading } = useThesisPicker()

  const consultations = useQuery({
    queryKey: ['consultations', myThesis?.id],
    queryFn: () => (myThesis?.id ? consultationApi.list(myThesis.id, { per_page: 50 }) : Promise.resolve(null)),
    enabled: Boolean(myThesis?.id),
  })

  const list = Array.isArray(consultations.data?.data)
    ? consultations.data.data
    : Array.isArray(consultations.data)
      ? consultations.data
      : []

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ConsultForm>({ resolver: zodResolver(consultSchema) })

  const create = useMutation({
    mutationFn: (data: ConsultForm) => {
      if (!myThesis?.id) throw new Error('Tidak ada skripsi aktif')
      return consultationApi.create(myThesis.id, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['consultations'] })
      queryClient.invalidateQueries({ queryKey: ['dashboard', 'student'] })
      reset()
      setShowForm(false)
    },
  })

  const approve = useMutation({
    mutationFn: (c: Consultation) => {
      if (!myThesis?.id) throw new Error('Tidak ada skripsi')
      return consultationApi.approve(myThesis.id, c.id)
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['consultations'] }),
  })

  if (thesesLoading) return <ListSkeleton count={4} label="Memuat data…" />

  if (!myThesis) {
    return (
      <Card>
        <CardContent className="py-14 text-center">
          <MessageSquarePlus className="mx-auto h-10 w-10 text-muted-foreground" />
          <p className="mt-4 text-lg font-semibold">Belum ada skripsi aktif</p>
          <p className="mt-1 text-sm text-muted-foreground">
            {isStudent ? 'Ajukan skripsi Anda terlebih dahulu untuk mulai bimbingan.' : 'Belum ada mahasiswa yang Anda bimbing.'}
          </p>
        </CardContent>
      </Card>
    )
  }

  const onSubmit = handleSubmit(async (data) => {
    setActionError(null)
    try {
      await create.mutateAsync(data)
    } catch (err) {
      setActionError(getErrorMessage(err, 'Gagal menyimpan bimbingan.'))
    }
  })

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">Bimbingan</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {isStudent ? 'Catatan bimbingan dengan dosen pembimbing Anda' : 'Catatan bimbingan mahasiswa Anda'}
          </p>
          {!isStudent && picker}
        </div>
        {isStudent && (
          <Button onClick={() => setShowForm((v) => !v)}>
            <MessageSquarePlus className="h-4 w-4" /> Catat Bimbingan
          </Button>
        )}
      </div>

      {showForm && (
        <Card>
          <CardHeader>
            <CardTitle className="text-base">Catat Bimbingan Baru</CardTitle>
          </CardHeader>
          <CardContent>
            <form onSubmit={onSubmit} noValidate className="space-y-4">
              {actionError && <Alert variant="danger">{actionError}</Alert>}
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div>
                  <Label htmlFor="date" required>Tanggal</Label>
                  <Input id="date" type="date" invalid={!!errors.date} {...register('date')} />
                  {errors.date && <p className="mt-1 text-xs text-danger">{errors.date.message}</p>}
                </div>
                <div>
                  <Label htmlFor="topic" required>Topik / Materi</Label>
                  <Input id="topic" placeholder="mis. Revisi Bab 2" invalid={!!errors.topic} {...register('topic')} />
                  {errors.topic && <p className="mt-1 text-xs text-danger">{errors.topic.message}</p>}
                </div>
              </div>
              <div>
                <Label htmlFor="notes">Catatan</Label>
                <Textarea id="notes" rows={3} placeholder="Hasil diskusi…" {...register('notes')} />
              </div>
              <div>
                <Label htmlFor="follow_up">Tindak Lanjut</Label>
                <Textarea id="follow_up" rows={2} placeholder="Pekerjaan selanjutnya…" {...register('follow_up')} />
              </div>
              <div className="flex gap-3">
                <Button type="submit" loading={create.isPending}>Simpan</Button>
                <Button type="button" variant="ghost" onClick={() => setShowForm(false)}>Batal</Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      <div className="space-y-3">
        {consultations.isLoading && <Spinner />}
        {list.map((c) => (
          <Card key={c.id}>
            <CardContent className="p-5">
              <div className="flex flex-wrap items-start justify-between gap-3">
                <div className="min-w-0">
                  <div className="flex items-center gap-2">
                    <p className="font-medium">{c.topic}</p>
                    {statusBadge(c.status)}
                  </div>
                  <p className="mt-1 text-xs text-muted-foreground">
                    {formatDate(c.date)}
                    {c.student_name && ` · ${c.student_name}`}
                    {c.supervisor_name && ` · ${c.supervisor_name}`}
                  </p>
                </div>
                {!isStudent && c.status === 'pending' && (
                  <Button size="sm" variant="success" onClick={() => approve.mutate(c)}>
                    <CheckCircle2 className="h-3.5 w-3.5" /> Setujui
                  </Button>
                )}
              </div>
              {c.notes && <p className="mt-3 text-sm leading-relaxed">{c.notes}</p>}
              {c.follow_up && (
                <p className="mt-2 text-sm text-muted-foreground">
                  <span className="font-medium">Tindak lanjut:</span> {c.follow_up}
                </p>
              )}
            </CardContent>
          </Card>
        ))}
        {!consultations.isLoading && list.length === 0 && (
          <p className="py-10 text-center text-sm text-muted-foreground">Belum ada catatan bimbingan.</p>
        )}
      </div>
    </div>
  )
}
