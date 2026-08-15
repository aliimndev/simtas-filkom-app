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
      <div>
        <p className="landing-eyebrow">Jadwal Ujian</p>
        <h1 className="mt-2 text-balance landing-heading text-2xl">
          Seminar & <span className="accent-text italic">sidang</span> 14 hari ke depan
        </h1>
      </div>

      {upcoming.isLoading ? (
        <ListSkeleton count={5} label="Memuat jadwal…" />
      ) : items.length === 0 ? (
        <div className="py-12 text-center">
          <CalendarDays className="mx-auto h-10 w-10 text-st-muted/40" />
          <p className="mt-3 landing-heading text-lg">Tidak ada <span className="accent-text italic">jadwal mendatang</span></p>
          <p className="mt-1 text-sm text-st-muted">Jadwal ujian akan muncul di sini.</p>
        </div>
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
                    <div className={`flex h-10 w-10 items-center justify-center rounded-full ${item.type === 'seminar' ? 'bg-(--st-accent-from)/10 text-(--st-accent-to)' : 'bg-success-50 text-success'}`}>
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
