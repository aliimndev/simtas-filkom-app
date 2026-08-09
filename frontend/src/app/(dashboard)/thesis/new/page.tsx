'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { ArrowLeft, BookOpen } from 'lucide-react'
import { useForm, useWatch } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { Textarea } from '@/components/ui/textarea'
import { FileDropzone } from '@/components/ui/file-dropzone'
import { thesisApi } from '@/lib/api/thesis-api'
import { academicYearApi } from '@/lib/api/user-api'
import { getErrorMessage } from '@/lib/utils/error'

const MAX_TITLE_CHARS = 500

// Hitung kata untuk frontend yang sinkron dengan aturan backend
// (backend: len(strings.Fields(...)) < 100 / < 10) — lihat thesis_usecase.go:140-148.
const wordCount = (s: string) => s.trim().split(/\s+/).filter(Boolean).length

const thesisSchema = z.object({
  title: z
    .string()
    .min(10, 'Judul minimal 10 karakter')
    .max(MAX_TITLE_CHARS, `Judul maksimal ${MAX_TITLE_CHARS} karakter`)
    .refine((val) => wordCount(val) >= 10, { message: 'Judul minimal 10 kata' }),
  abstract: z
    .string()
    .refine((val) => wordCount(val) >= 100, { message: 'Abstrak minimal 100 kata' }),
  file: z.instanceof(File, { message: 'Draft proposal wajib diunggah' }),
  field_of_study: z.string().min(1, 'Bidang keahlian wajib diisi'),
  thesis_type: z.enum(['skripsi', 'tugas_akhir'], {
    required_error: 'Tipe skripsi wajib dipilih',
  }),
  academic_year_id: z.string().min(1, 'Pilih tahun akademik'),
})

type ThesisForm = z.infer<typeof thesisSchema>

export default function NewThesisPage() {
  const router = useRouter()
  const queryClient = useQueryClient()
  const years = useQuery({ queryKey: ['academic-years'], queryFn: academicYearApi.list })

  const {
    register,
    handleSubmit,
    control,
    setValue,
    formState: { errors },
  } = useForm<ThesisForm>({ resolver: zodResolver(thesisSchema) })

  const draftFile = useWatch({ control, name: 'file' })

  const create = useMutation({
    mutationFn: (data: ThesisForm) => thesisApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['theses'] })
      queryClient.invalidateQueries({ queryKey: ['dashboard', 'student'] })
      router.push('/thesis')
    },
  })

  const onSubmit = handleSubmit(async (data) => {
    try {
      await create.mutateAsync(data)
    } catch {
      /* error shown via mutation state */
    }
  })

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <div>
        <Link
          href="/dashboard"
          className="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" /> Kembali ke Dashboard
        </Link>
        <h1 className="text-2xl font-bold">Ajukan Judul Skripsi</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Lengkapi form di bawah untuk mengajukan judul. Kaprodi akan mereview pengajuan Anda.
        </p>
      </div>

      {create.isError && (
        <Alert variant="danger">{getErrorMessage(create.error, 'Gagal mengajukan skripsi.')}</Alert>
      )}
      {create.isSuccess && (
        <Alert variant="success">
          Pengajuan berhasil dikirim! Anda akan diarahkan ke halaman skripsi Anda.
        </Alert>
      )}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <BookOpen className="h-5 w-5 text-primary" /> Form Pengajuan
          </CardTitle>
          <CardDescription>Semua kolom wajib diisi</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} noValidate className="space-y-4">
            <div>
              <Label htmlFor="title" required>Judul Skripsi</Label>
              <Input id="title" placeholder="Judul lengkap skripsi Anda" invalid={!!errors.title} {...register('title')} />
              {errors.title && <p className="mt-1 text-xs text-danger">{errors.title.message}</p>}
            </div>

            <div>
              <Label htmlFor="abstract" required>Abstrak</Label>
              <Textarea
                id="abstract"
                rows={6}
                placeholder="Ringkasan singkat tentang latar belakang, tujuan, dan metode penelitian…"
                invalid={!!errors.abstract}
                {...register('abstract')}
              />
              {errors.abstract && <p className="mt-1 text-xs text-danger">{errors.abstract.message}</p>}
            </div>

            <div>
              <Label htmlFor="draft_file" required>Draft Proposal Skripsi</Label>
              <FileDropzone
                id="draft_file"
                value={draftFile}
                invalid={!!errors.file}
                onChange={(file) => setValue('file', file as File, { shouldValidate: true })}
              />
              <p className="mt-1 text-xs text-muted-foreground">
                Upload draft proposal skripsi dalam format PDF. Maksimal 10 MB.
              </p>
              {errors.file && <p className="mt-1 text-xs text-danger">{errors.file.message}</p>}
            </div>

            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <Label htmlFor="field_of_study" required>Bidang Keahlian</Label>
                <Input id="field_of_study" placeholder="mis. Rekayasa Perangkat Lunak" invalid={!!errors.field_of_study} {...register('field_of_study')} />
                {errors.field_of_study && <p className="mt-1 text-xs text-danger">{errors.field_of_study.message}</p>}
              </div>
              <div>
                <Label htmlFor="academic_year_id" required>Tahun Akademik</Label>
                <Select
                  id="academic_year_id"
                  invalid={!!errors.academic_year_id}
                  {...register('academic_year_id')}
                >
                  <option value="">Pilih tahun akademik…</option>
                  {years.data?.map((y) => (
                    <option key={y.id} value={y.id}>
                      {y.name}
                      {y.is_active ? ' (aktif)' : ''}
                    </option>
                  ))}
                </Select>
                {errors.academic_year_id && <p className="mt-1 text-xs text-danger">{errors.academic_year_id.message}</p>}
              </div>
            </div>

            <div>
              <Label htmlFor="thesis_type" required>Tipe Skripsi</Label>
              <Select
                id="thesis_type"
                invalid={!!errors.thesis_type}
                {...register('thesis_type')}
              >
                <option value="">Pilih tipe skripsi…</option>
                <option value="skripsi">Skripsi</option>
                <option value="tugas_akhir">Tugas Akhir</option>
              </Select>
              {errors.thesis_type && (
                <p className="mt-1 text-xs text-danger">{errors.thesis_type.message}</p>
              )}
            </div>

            <div className="flex gap-3 pt-2">
              <Button type="submit" loading={create.isPending}>
                Ajukan Skripsi
              </Button>
              <Button type="button" variant="outline" onClick={() => window.history.back()}>
                Batal
              </Button>
            </div>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}
