'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { MessageSquarePlus } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Spinner } from '@/components/ui/spinner'
import { ListSkeleton } from '@/components/ui/skeleton'
import { consultationApi } from '@/lib/api/consultation-api'
import { getErrorMessage } from '@/lib/utils/error'
import { useAuthStore } from '@/lib/stores/auth-store'
import { useThesisPicker } from '@/lib/hooks/use-thesis-picker'
import { ConsultationForm } from '@/components/features/supervision-page/consultation-form'
import { ConsultationRow } from '@/components/features/supervision-page/consultation-row'
import type { Consultation } from '@/types/consultation'

export default function SupervisionPage() {
  const user = useAuthStore((s) => s.user)
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

  const create = useMutation({
    mutationFn: (data: { date: string; topic: string; notes?: string; follow_up?: string }) => {
      if (!myThesis?.id) throw new Error('Tidak ada skripsi aktif')
      return consultationApi.create(myThesis.id, data)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['consultations'] })
      queryClient.invalidateQueries({ queryKey: ['dashboard', 'student'] })
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

  const onSubmit = async (data: { date: string; topic: string; notes?: string; follow_up?: string }) => {
    setActionError(null)
    try {
      await create.mutateAsync(data)
    } catch (err) {
      setActionError(getErrorMessage(err, 'Gagal menyimpan bimbingan.'))
    }
  }

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
        <ConsultationForm
          onClose={() => setShowForm(false)}
          actionError={actionError}
          isSubmitting={create.isPending}
          onSubmit={onSubmit}
        />
      )}

      <div className="space-y-3">
        {consultations.isLoading && <Spinner />}
        {list.map((c) => (
          <ConsultationRow
            key={c.id}
            consultation={c}
            isStudent={isStudent}
            onApprove={() => approve.mutate(c)}
          />
        ))}
        {!consultations.isLoading && list.length === 0 && (
          <p className="py-10 text-center text-sm text-muted-foreground">Belum ada catatan bimbingan.</p>
        )}
      </div>
    </div>
  )
}
