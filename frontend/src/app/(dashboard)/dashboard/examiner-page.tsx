'use client'

import { useQuery } from '@tanstack/react-query'
import { CalendarDays, ClipboardCheck, Star } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard, ViewAllLink } from '@/components/features/dashboard/stat-card'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { formatDateTime } from '@/lib/utils/date'
import type { ExaminerAssignment } from '@/types/dashboard'

function greeting(): string {
  const h = new Date().getHours()
  if (h < 11) return 'Selamat pagi'
  if (h < 15) return 'Selamat siang'
  if (h < 18) return 'Selamat sore'
  return 'Selamat malam'
}

function AssignmentRow({ a }: { a: ExaminerAssignment }) {
  return (
    <div className="flex items-center gap-3 rounded-lg border border-border p-3">
      <div className="flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary">
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
  const { user } = useAuthStore()
  const dash = useQuery({
    queryKey: ['dashboard', 'examiner'],
    queryFn: dashboardApi.examiner,
    refetchInterval: 5 * 60 * 1000,
  })

  const d = dash.data

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">{greeting()}, {user?.full_name?.split(' ')[0]} 👋</h1>
        <p className="mt-1 text-sm text-muted-foreground">{roleLabel(user?.role)} — NIDN {user?.nim_nidn ?? '—'}</p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard title="Jadwal Mendatang" value={d?.upcoming_assignments?.length ?? '—'} icon={CalendarDays} href="/schedules" iconClass="bg-primary-50 text-primary" />
        <StatCard title="Nilai Belum Diinput" value={d?.pending_scores?.length ?? '—'} icon={ClipboardCheck} href="/defenses" iconClass="bg-warning-50 text-warning" />
        <StatCard title="Total Sudah Dinilai" value={d?.scoring_history?.length ?? '—'} icon={Star} href="/defenses" iconClass="bg-success-50 text-success" />
      </div>

      <Card>
        <CardHeader className="flex-row items-center justify-between space-y-0">
          <div>
            <CardTitle>Nilai Belum Diinput</CardTitle>
            <CardDescription>Ujian yang menunggu penilaian Anda</CardDescription>
          </div>
          <ViewAllLink href="/defenses" label="Lihat Jadwal" />
        </CardHeader>
        <CardContent className="space-y-3">
          {dash.isLoading && <Skeleton className="h-24 w-full" />}
          {d?.pending_scores?.map((a) => <AssignmentRow key={a.id} a={a} />)}
          {!dash.isLoading && (!d?.pending_scores || d.pending_scores.length === 0) && (
            <p className="text-sm text-muted-foreground">Tidak ada nilai tertunda. 🎉</p>
          )}
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <CardTitle>Jadwal Mendatang</CardTitle>
          <CardDescription>Ujian yang akan datang</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {dash.isLoading && <Skeleton className="h-24 w-full" />}
          {d?.upcoming_assignments?.map((a) => <AssignmentRow key={a.id} a={a} />)}
          {!dash.isLoading && (!d?.upcoming_assignments || d.upcoming_assignments.length === 0) && (
            <p className="text-sm text-muted-foreground">Tidak ada jadwal mendatang.</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
