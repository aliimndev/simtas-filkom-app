'use client'

import { useQuery } from '@tanstack/react-query'
import { GraduationCap, CalendarDays, MapPin, Star } from 'lucide-react'
import { useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { ListSkeleton } from '@/components/ui/skeleton'
import { seminarApi } from '@/lib/api/seminar-api'
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

export default function SeminarsPage() {
  const [status, setStatus] = useState('')

  const seminars = useQuery({
    queryKey: ['seminars', status],
    queryFn: () => seminarApi.list({ per_page: 50, status: status || undefined }),
  })

  const list = seminars.data?.data ?? []

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">Seminar</h1>
        <p className="mt-1 text-sm text-muted-foreground">Jadwal dan hasil seminar proposal / kemajuan</p>
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

      {seminars.isLoading ? (
        <ListSkeleton count={5} label="Memuat seminar…" />
      ) : (
        <div className="space-y-3">
          {list.map((sem) => (
            <Card key={sem.id}>
              <CardContent className="flex flex-wrap items-center justify-between gap-3 p-5">
                <div className="flex min-w-0 items-center gap-3">
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl bg-primary-50 text-primary">
                    <GraduationCap className="h-5 w-5" />
                  </div>
                  <div className="min-w-0">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="font-medium">{sem.student?.full_name ?? 'Mahasiswa'}</p>
                      {statusBadge(sem.status)}
                      <Badge variant="muted">Seminar {sem.stage}</Badge>
                    </div>
                    <p className="mt-0.5 truncate text-sm text-muted-foreground">{sem.thesis_title ?? '—'}</p>
                    <div className="mt-1.5 flex flex-wrap items-center gap-x-4 gap-y-1 text-xs text-muted-foreground">
                      {sem.scheduled_at && (
                        <span className="inline-flex items-center gap-1">
                          <CalendarDays className="h-3.5 w-3.5" /> {formatDateTime(sem.scheduled_at)}
                        </span>
                      )}
                      {sem.room && (
                        <span className="inline-flex items-center gap-1">
                          <MapPin className="h-3.5 w-3.5" /> {sem.room}
                        </span>
                      )}
                      {sem.average_score != null && (
                        <span className="inline-flex items-center gap-1 text-warning">
                          <Star className="h-3.5 w-3.5" /> {sem.average_score.toFixed(2)}
                        </span>
                      )}
                    </div>
                  </div>
                </div>
                {sem.examiners && sem.examiners.length > 0 && (
                  <div className="hidden text-right text-xs text-muted-foreground md:block">
                    <p className="font-medium text-foreground">Penguji</p>
                    {sem.examiners.map((e) => (
                      <p key={e.id}>{e.full_name}</p>
                    ))}
                  </div>
                )}
              </CardContent>
            </Card>
          ))}
          {list.length === 0 && (
            <p className="py-10 text-center text-sm text-muted-foreground">Tidak ada seminar.</p>
          )}
        </div>
      )}
    </div>
  )
}
