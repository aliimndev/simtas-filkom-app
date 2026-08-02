'use client'

import { useQuery } from '@tanstack/react-query'
import { BookOpen } from 'lucide-react'
import Link from 'next/link'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Spinner } from '@/components/ui/spinner'
import { thesisApi } from '@/lib/api/thesis-api'
import { useAuthStore } from '@/lib/stores/auth-store'

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

export default function MyThesisPage() {
  const { user } = useAuthStore()
  const theses = useQuery({
    queryKey: ['theses', 'mine'],
    queryFn: () => thesisApi.list({ per_page: 1 }),
    enabled: Boolean(user),
  })

  const list = theses.data?.data ?? []
  const thesis = list[0]

  if (theses.isLoading) return <Spinner label="Memuat skripsi…" />

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
        <Button asChild variant="outline" size="sm">
          <Link href="/thesis/new">Ajukan Judul Baru</Link>
        </Button>
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
    </div>
  )
}
