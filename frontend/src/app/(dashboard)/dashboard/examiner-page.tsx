'use client'

import { useQuery } from '@tanstack/react-query'
import { CalendarDays, ClipboardCheck, Star } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard, ViewAllLink } from '@/components/features/dashboard/stat-card'
import { DashboardHeader } from '@/components/features/dashboard/dashboard-header'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { formatDateTime } from '@/lib/utils/date'
import type { ExaminerAssignment } from '@/types/dashboard'

const ROW = 'flex items-center gap-3 rounded-md border border-border bg-muted/40 px-4 py-3'

function AssignmentRow({ a }: { a: ExaminerAssignment }) {
  return (
    <div className={ROW}>
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-primary-50 text-primary">
        <ClipboardCheck className="h-4 w-4" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium">{a.student_name ?? a.thesis_title}</p>
        <p className="text-xs text-muted-foreground">
          {a.type === 'seminar' ? 'Seminar' : 'Sidang'} · {formatDateTime(a.scheduled_at)}
          {a.room ? ` · ${a.room}` : ''}
        </p>
      </div>
      {a.has_scored ? <Badge variant="success">Dinilai</Badge> : <Badge variant="warning">Belum dinilai</Badge>}
    </div>
  )
}

export default function ExaminerDashboardPage() {
  const user = useAuthStore((s) => s.user)
  const dash = useQuery({
    queryKey: ['dashboard', 'examiner'],
    queryFn: dashboardApi.examiner,
    refetchInterval: 5 * 60 * 1000,
  })

  const d = dash.data

  return (
    <div className="space-y-6">
      <DashboardHeader
        name={user?.full_name ?? 'Penguji'}
        subtitle={`${roleLabel(user?.role)} · NIDN ${user?.nim_nidn ?? '—'}`}
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard title="Jadwal Mendatang" value={d?.upcoming_assignments?.length ?? '—'} icon={CalendarDays} href="/schedules" />
        <StatCard title="Nilai Belum Diinput" value={d?.pending_scores?.length ?? '—'} icon={ClipboardCheck} href="/defenses" />
        <StatCard title="Total Sudah Dinilai" value={d?.scoring_history?.length ?? '—'} icon={Star} href="/defenses" />
      </div>

      <Card>
        <CardHeader className="flex-row items-center justify-between space-y-0">
          <div>
            <CardTitle>Nilai Belum Diinput</CardTitle>
            <p className="mt-0.5 text-sm text-muted-foreground">Ujian yang menunggu penilaian Anda</p>
          </div>
          <ViewAllLink href="/defenses" label="Lihat Jadwal" />
        </CardHeader>
        <CardContent className="space-y-2.5">
          {dash.isLoading && <Skeleton className="h-20 w-full" />}
          {d?.pending_scores?.map((a) => <AssignmentRow key={a.id} a={a} />)}
          {!dash.isLoading && (!d?.pending_scores || d.pending_scores.length === 0) && (
            <p className="py-2 text-sm text-muted-foreground">Tidak ada nilai tertunda.</p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader className="flex-row items-center justify-between space-y-0">
          <div>
            <CardTitle>Jadwal Mendatang</CardTitle>
            <p className="mt-0.5 text-sm text-muted-foreground">Ujian yang akan datang</p>
          </div>
          <ViewAllLink href="/schedules" />
        </CardHeader>
        <CardContent className="space-y-2.5">
          {dash.isLoading && <Skeleton className="h-20 w-full" />}
          {d?.upcoming_assignments?.map((a) => <AssignmentRow key={a.id} a={a} />)}
          {!dash.isLoading && (!d?.upcoming_assignments || d.upcoming_assignments.length === 0) && (
            <p className="py-2 text-sm text-muted-foreground">Tidak ada jadwal mendatang.</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
