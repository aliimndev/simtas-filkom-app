'use client'

import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

const consultSchema = z.object({
  date: z.string().min(1, 'Tanggal wajib diisi'),
  topic: z.string().min(3, 'Topik wajib diisi'),
  notes: z.string().optional(),
  follow_up: z.string().optional(),
})

type ConsultForm = z.infer<typeof consultSchema>

/** Form pencatatan bimbingan baru — skema + state form hidup di komponen ini. */
export function ConsultationForm({
  onClose,
  actionError,
  isSubmitting,
  onSubmit,
}: {
  onClose: () => void
  actionError: string | null
  isSubmitting: boolean
  onSubmit: (data: ConsultForm) => Promise<void>
}) {
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ConsultForm>({ resolver: zodResolver(consultSchema) })

  // Catatan: form di-unmount saat ditutup (halaman merender kondisional),
  // sehingga state tidak perlu di-reset manual.

  return (
    <Card>
      <CardHeader>
        <CardTitle className="text-base">Catat Bimbingan Baru</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} noValidate className="space-y-4">
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
            <Button type="submit" loading={isSubmitting}>Simpan</Button>
            <Button type="button" variant="ghost" onClick={onClose}>Batal</Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
