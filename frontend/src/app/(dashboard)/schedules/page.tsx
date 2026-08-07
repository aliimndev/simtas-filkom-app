'use client'

import { useQuery } from '@tanstack/react-query'
import { CalendarDays, GraduationCap, ClipboardCheck, MapPin } from 'lucide-react'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { ListSkeleton } from '@/components/ui/skeleton'
import { defenseApi } from '@/lib/api/defense-api'
import { formatDateTime } from '@/lib/utils/date'
import { flattenSchedules, type UpcomingSchedules } from '@/types/dashboard'

type FlatItem = ReturnType<typeof flattenSchedules>[number]

export default function SchedulesPage() {
  const upcoming = useQuery<UpcomingSchedules>({
    queryKey: ['schedules', 'upcoming'],
    queryFn: defenseApi.upcoming,
    refetchInterval: 5 * 60 * 1000,
  })

  const items = flattenSchedules(upcoming.data)

  const grouped = items.reduce<Record<string, FlatItem[]>>((acc, item) => {
    const key = item.scheduled_at ? new Date(item.scheduled_at).toDateString() : 'Lainnya'
    if (!acc[key]) acc[key] = []
    acc[key].push(item)
    return acc
  }, {})

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary-50 text-primary">
          <CalendarDays className="h-5 w-5" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">Jadwal Ujian</h1>
          <p className="text-sm text-muted-foreground">Seminar & sidang 14 hari ke depan</p>
        </div>
      </div>

      {upcoming.isLoading ? (
        <ListSkeleton count={5} label="Memuat jadwal…" />
      ) : items.length === 0 ? (
        <p className="py-10 text-center text-sm text-muted-foreground">Tidak ada jadwal mendatang.</p>
      ) : (
        Object.entries(grouped).map(([dateKey, dayItems]) => (
          <div key={dateKey}>
            <p className="mb-2 text-sm font-semibold uppercase tracking-wide text-muted-foreground">
              {formatDateTime(dateKey).split(',')[0]}
            </p>
            <div className="space-y-2">
              {dayItems.map((item) => (
                <Card key={`${item.type}-${item.id}`}>
                  <CardContent className="flex flex-wrap items-center gap-3 p-4">
                    <div className={`flex h-10 w-10 items-center justify-center rounded-lg ${item.type === 'seminar' ? 'bg-primary-50 text-primary' : 'bg-success-50 text-success'}`}>
                      {item.type === 'seminar' ? <GraduationCap className="h-5 w-5" /> : <ClipboardCheck className="h-5 w-5" />}
                    </div>
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <p className="font-medium">{item.student_name ?? 'Mahasiswa'}</p>
                        <Badge variant={item.type === 'seminar' ? 'primary' : 'success'}>
                          {item.type === 'seminar' ? 'Seminar' : 'Sidang'}
                        </Badge>
                      </div>
                      <p className="mt-0.5 line-clamp-1 text-sm text-muted-foreground">{item.thesis_title ?? ''}</p>
                    </div>
                    <div className="flex items-center gap-3 text-sm text-muted-foreground">
                      <span className="inline-flex items-center gap-1">
                        <CalendarDays className="h-3.5 w-3.5" /> {formatDateTime(item.scheduled_at)}
                      </span>
                      {item.room && (
                        <span className="inline-flex items-center gap-1">
                          <MapPin className="h-3.5 w-3.5" /> {item.room}
                        </span>
                      )}
                    </div>
                  </CardContent>
                </Card>
              ))}
            </div>
          </div>
        ))
      )}
    </div>
  )
}
