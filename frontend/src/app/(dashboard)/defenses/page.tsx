'use client'

import { useQuery } from '@tanstack/react-query'
import { ClipboardCheck, CalendarDays, MapPin, Star } from 'lucide-react'
import { useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { ListSkeleton } from '@/components/ui/skeleton'
import { defenseApi } from '@/lib/api/defense-api'
import { formatDateTime } from '@/lib/utils/date'

function statusBadge(status: string) {
  switch (status) {
    case 'passed':
      return <Badge variant="success">Lulus</Badge>
    case 'failed':
      return <Badge variant="danger">Tidak Lulus</Badge>
    case 'in_revision':
      return <Badge variant="warning">Revisi</Badge>
    case 'scheduled':
      return <Badge variant="primary">Terjadwal</Badge>
    default:
      return <Badge variant="muted">Diajukan</Badge>
  }
}

export default function DefensesPage() {
  const [status, setStatus] = useState('')

  const defenses = useQuery({
    queryKey: ['defenses', status],
    queryFn: () => defenseApi.list({ per_page: 50, status: status || undefined }),
  })

  const list = defenses.data?.data ?? []

  return (
    <div className="space-y-6">
      <div>
        <p className="landing-eyebrow">Sidang Tugas Akhir</p>
        <h1 className="mt-2 text-balance landing-heading text-2xl">
          Jadwal dan hasil <span className="accent-text italic">sidang</span>
        </h1>
        <p className="mt-1.5 text-sm text-st-muted">Jadwal dan hasil sidang tugas akhir</p>
      </div>

      <div className="flex flex-wrap gap-2">
        {['', 'submitted', 'scheduled', 'in_revision', 'passed', 'failed'].map((s) => (
          <Button
            key={s}
            size="sm"
            variant={status === s ? 'primary' : 'outline'}
            onClick={() => setStatus(s)}
          >
            {s === '' ? 'Semua' : s.replace(/_/g, ' ')}
          </Button>
        ))}
      </div>

      {defenses.isLoading ? (
        <ListSkeleton count={5} label="Memuat sidang…" />
      ) : (
        <div className="space-y-3">
          {list.map((d) => (
            <Card key={d.id}>
              <CardContent className="flex flex-wrap items-center justify-between gap-3 p-5">
                <div className="flex min-w-0 items-center gap-3">
                  <div className="accent-ring flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-success-50 text-success">
                    <ClipboardCheck className="h-5 w-5" />
                  </div>
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-medium">{d.student?.full_name ?? 'Mahasiswa'}</p>
                      {statusBadge(d.status)}
                    </div>
                    <p className="mt-0.5 truncate text-sm text-muted-foreground">{d.thesis_title ?? '—'}</p>
                    <div className="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
                      {d.scheduled_at && (
                        <span className="inline-flex items-center gap-1">
                          <CalendarDays className="h-3.5 w-3.5" /> {formatDateTime(d.scheduled_at)}
                        </span>
                      )}
                      {d.room && (
                        <span className="inline-flex items-center gap-1">
                          <MapPin className="h-3.5 w-3.5" /> {d.room}
                        </span>
                      )}
                      {d.average_score != null && (
                        <span className="inline-flex items-center gap-1 text-warning">
                          <Star className="h-3.5 w-3.5" /> {d.average_score.toFixed(2)}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                {d.examiners && d.examiners.length > 0 && (
                  <div className="hidden text-right text-xs text-muted-foreground md:block">
                    <p className="font-medium text-foreground">Penguji</p>
                    {d.examiners.map((e) => (
                      <p key={e.id}>{e.full_name}</p>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
          {list.length === 0 && (
            <div className="py-12 text-center">
              <ClipboardCheck className="mx-auto h-10 w-10 text-st-muted/40" />
              <p className="mt-3 landing-heading text-lg">Tidak ada <span className="accent-text italic">sidang</span></p>
              <p className="mt-1 text-sm text-st-muted">Jadwal sidang akan muncul di sini.</p>
            </div>
          )}
        </div>
      )}
    </div>
  )
}
