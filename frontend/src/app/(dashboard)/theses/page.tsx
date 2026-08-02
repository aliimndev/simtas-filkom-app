'use client'

import { useQuery } from '@tanstack/react-query'
import { BookOpen } from 'lucide-react'
import { useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { thesisApi } from '@/lib/api/thesis-api'
import { formatDate } from '@/lib/utils/date'
import type { ThesisStatus } from '@/types/thesis'

function statusVariant(status: string) {
  switch (status) {
    case 'graduated':
      return 'success'
    case 'cancelled':
    case 'rejected':
      return 'danger'
    case 'submitted':
      return 'warning'
    case 'defense_done':
      return 'success'
    default:
      return 'primary'
  }
}

export default function ThesesPage() {
  const [q, setQ] = useState('')
  const [status, setStatus] = useState('')
  const [page, setPage] = useState(1)

  const theses = useQuery({
    queryKey: ['theses', 'admin', q, status, page],
    queryFn: () => thesisApi.list({ q: q || undefined, status: (status || undefined) as ThesisStatus | undefined, page, per_page: 10 }),
  })

  const list = theses.data?.data ?? []
  const total = theses.data?.meta.total ?? 0

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Daftar Skripsi</h1>
        <p className="mt-1 text-sm text-muted-foreground">Kelola pengajuan dan status skripsi mahasiswa</p>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <Input
          placeholder="Cari judul / mahasiswa…"
          value={q}
          onChange={(e) => {
            setQ(e.target.value)
            setPage(1)
          }}
          className="max-w-xs"
        />
        <Select value={status} onChange={(e) => { setStatus(e.target.value); setPage(1) }} className="w-44">
          <option value="">Semua status</option>
          <option value="submitted">Diajukan</option>
          <option value="approved">Disetujui</option>
          <option value="in_progress">Bimbingan</option>
          <option value="seminar_ready">Siap Seminar</option>
          <option value="defense_ready">Siap Sidang</option>
          <option value="graduated">Lulus</option>
          <option value="cancelled">Dibatalkan</option>
        </Select>
      </div>

      {theses.isLoading ? (
        <Spinner />
      ) : (
        <div className="space-y-3">
          {list.map((t) => (
            <Card key={t.id}>
              <CardContent className="flex flex-wrap items-center justify-between gap-3 p-5">
                <div className="flex min-w-0 items-center gap-3">
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary-50 text-primary">
                    <BookOpen className="h-5 w-5" />
                  </div>
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-medium">{t.student?.full_name ?? '—'}</p>
                      <Badge variant={statusVariant(t.status)}>{t.status.replace(/_/g, ' ')}</Badge>
                    </div>
                    <p className="mt-0.5 line-clamp-1 text-sm text-muted-foreground">{t.title}</p>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Diajukan {formatDate(t.submitted_at)}
                      {t.supervisors?.length ? ` · Pembimbing: ${t.supervisors.map((s) => s.full_name).join(', ')}` : ''}
                    </p>
                  </div>
                </div>
                <div className="flex gap-2">
                  {t.status === 'submitted' && (
                    <Button asChild size="sm">
                      <a href={`/theses/${t.id}/review`}>Review</a>
                    </Button>
                  )}
                  {(t.status === 'approved' || t.status === 'in_progress') && (
                    <Button asChild size="sm" variant="outline">
                      <a href={`/theses/${t.id}/assign`}>Atur Pembimbing</a>
                    </Button>
                  )}
                </div>
              </CardContent>
            </Card>
          ))}
          {list.length === 0 && <p className="py-10 text-center text-sm text-muted-foreground">Tidak ada skripsi ditemukan.</p>}
        </div>
      )}

      {total > 10 && (
        <div className="flex items-center justify-between pt-2">
          <p className="text-sm text-muted-foreground">
            Total {total} skripsi · Halaman {page}
          </p>
          <div className="flex gap-2">
            <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>
              Sebelumnya
            </Button>
            <Button size="sm" variant="outline" disabled={page >= Math.ceil(total / 10)} onClick={() => setPage((p) => p + 1)}>
              Berikutnya
            </Button>
          </div>
        </div>
      )}
    </div>
  )
}
