'use client'

import { useEffect } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { PencilLine } from 'lucide-react'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Dialog } from '@/components/ui/dialog'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'

const submitSchema = z.object({
  requested_title: z
    .string()
    .min(1, 'Judul baru wajib diisi')
    .max(500, 'Judul maksimal 500 karakter')
    .refine((v) => v.trim().split(/\s+/).filter(Boolean).length >= 10, 'Judul minimal 10 kata'),
  reason: z.string().optional(),
})

type SubmitForm = z.infer<typeof submitSchema>

/** Dialog pengajuan perubahan judul — form + validasinya hidup di sini. */
export function TitleChangeFormDialog({
  open,
  onClose,
  currentTitle,
  actionError,
  isSubmitting,
  onSubmit,
}: {
  open: boolean
  onClose: () => void
  currentTitle: string
  actionError: string | null
  isSubmitting: boolean
  onSubmit: (data: SubmitForm) => Promise<void>
}) {
  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<SubmitForm>({ resolver: zodResolver(submitSchema) })

  // Form di-reset setiap dialog ditutup agar tidak menyisakan input lama.
  useEffect(() => {
    if (!open) reset()
  }, [open, reset])

  return (
    <Dialog
      open={open}
      onClose={onClose}
      title="Ajukan Perubahan Judul"
      description="Perubahan judul akan diproses oleh Dosen Pembimbing Anda."
    >
      <form onSubmit={handleSubmit(onSubmit)} noValidate className="space-y-4">
        {actionError && <Alert variant="danger">{actionError}</Alert>}
        <div>
          <Label>Judul Saat Ini</Label>
          <Input value={currentTitle} readOnly className="bg-muted/60 text-muted-foreground" />
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
          <Button type="button" variant="ghost" onClick={onClose}>Batal</Button>
          <Button type="submit" loading={isSubmitting}>
            <PencilLine className="h-4 w-4" /> Ajukan
          </Button>
        </div>
      </form>
    </Dialog>
  )
}
