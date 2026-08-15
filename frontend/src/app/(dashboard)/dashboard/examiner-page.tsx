'use client'

import { useQuery } from '@tanstack/react-query'
import { CalendarDays, ClipboardCheck, Star } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard, ViewAllLink } from '@/components/features/dashboard/stat-card'
import { DashboardHeader } from '@/components/features/dashboard/dashboard-header'
import { DashboardHeroBand } from '@/components/features/dashboard/dashboard-hero-band'
import { StatusBadge } from '@/components/features/dashboard/status-badge'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { formatDateTime } from '@/lib/utils/date'
import type { ExaminerAssignment } from '@/types/dashboard'

function AssignmentRow({ a }: { a: ExaminerAssignment }) {
  return (
    <div className="flex items-center gap-3 rounded-xl border border-st-stroke bg-st-surface-hi/60 px-4 py-3">
      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-(--st-accent-from)/10 text-(--st-accent-to)">
        <ClipboardCheck className="h-4 w-4" />
      </div>
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-st-text">{a.student_name ?? a.thesis_title}</p>
        <p className="text-xs text-st-muted">
          {a.type === 'seminar' ? 'Seminar' : 'Sidang'} · {formatDateTime(a.scheduled_at)}
          {a.room ? ` · ${a.room}` : ''}
        </p>
      </div>
      {a.has_scored ? <StatusBadge variant="completed" label="Dinilai" /> : <StatusBadge variant="pending" label="Belum dinilai" />}
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
    <div className="space-y-8">
      <DashboardHeroBand>
        <DashboardHeader
          name={user?.full_name ?? 'Penguji'}
          subtitle={`${roleLabel(user?.role)} · NIDN ${user?.nim_nidn ?? '—'}`}
        />
      </DashboardHeroBand>

      {/* Stat cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard title="Jadwal Mendatang" value={d?.upcoming_assignments?.length ?? '—'} icon={CalendarDays} href="/schedules" />
        <StatCard title="Nilai Belum Diinput" value={d?.pending_scores?.length ?? '—'} icon={ClipboardCheck} href="/defenses" />
        <StatCard title="Total Sudah Dinilai" value={d?.scoring_history?.length ?? '—'} icon={Star} href="/defenses" />
      </div>

      {/* Pending scores */}
      <Card>
        <CardContent className="p-5 sm:p-6">
          <div className="mb-4 flex items-start justify-between gap-3">
            <div>
              <h2 className="landing-heading text-lg">Nilai <span className="accent-text italic">Belum Diinput</span></h2>
              <p className="mt-0.5 text-[13px] text-st-muted">Ujian yang menunggu penilaian Anda</p>
            </div>
            <ViewAllLink href="/defenses" label="Lihat Jadwal" />
          </div>
          <div className="space-y-2">
            {dash.isLoading && <Skeleton className="h-20 w-full" />}
            {d?.pending_scores?.map((a) => <AssignmentRow key={a.id} a={a} />)}
            {!dash.isLoading && (!d?.pending_scores || d.pending_scores.length === 0) && (
              <p className="py-2 text-sm text-st-muted">Tidak ada nilai tertunda.</p>
            )}
          </div>
        </CardContent>
      </Card>

      {/* Upcoming assignments */}
      <Card>
        <CardContent className="p-5 sm:p-6">
          <div className="mb-4 flex items-start justify-between gap-3">
            <div>
              <h2 className="landing-heading text-lg">Jadwal <span className="accent-text italic">Mendatang</span></h2>
              <p className="mt-0.5 text-[13px] text-st-muted">Ujian yang akan datang</p>
            </div>
            <ViewAllLink href="/schedules" />
          </div>
          <div className="space-y-2">
            {dash.isLoading && <Skeleton className="h-20 w-full" />}
            {d?.upcoming_assignments?.map((a) => <AssignmentRow key={a.id} a={a} />)}
            {!dash.isLoading && (!d?.upcoming_assignments || d.upcoming_assignments.length === 0) && (
              <p className="py-2 text-sm text-st-muted">Tidak ada jadwal mendatang.</p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

