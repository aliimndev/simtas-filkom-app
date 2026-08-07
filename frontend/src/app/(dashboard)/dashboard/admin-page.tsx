'use client'

import { useQuery } from '@tanstack/react-query'
import { Activity, Award, BookOpen, CalendarDays, ClipboardCheck, GraduationCap, Users } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard, ViewAllLink } from '@/components/features/dashboard/stat-card'
import { DashboardHeader } from '@/components/features/dashboard/dashboard-header'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { cn } from '@/lib/utils/cn'
import { flattenSchedules } from '@/types/dashboard'
import { formatDateTime } from '@/lib/utils/date'

const ROW = 'flex items-center justify-between gap-3 rounded-md border border-border bg-muted/40 px-4 py-3 text-sm'

export default function AdminDashboardPage() {
  const user = useAuthStore((s) => s.user)
  const summary = useQuery({
    queryKey: ['dashboard', 'summary'],
    queryFn: () => dashboardApi.summary(),
    refetchInterval: 5 * 60 * 1000,
  })
  const operational = useQuery({
    queryKey: ['dashboard', 'operational'],
    queryFn: () => dashboardApi.operational(),
    refetchInterval: 5 * 60 * 1000,
  })

  const s = summary.data
  const op = operational.data
  const schedules = flattenSchedules(op?.upcoming_schedules)
  const pending = op?.pending_actions
  const byStatus = s?.by_status ?? []

  const totalPending =
    pending
      ? pending.pending_title_reviews + pending.pending_document_reviews + pending.pending_seminars + pending.pending_defenses
      : 0

  return (
    <div className="space-y-6">
      <DashboardHeader
        name={user?.full_name ?? 'Admin'}
        subtitle={`Ringkasan operasional fakultas · ${roleLabel(user?.role)}`}
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard title="Mahasiswa Aktif" value={s?.academic_summary.total_active ?? '—'} icon={Users} href="/admin/users?role=mahasiswa" />
        <StatCard title="Total Lulus" value={s?.academic_summary.total_graduated ?? '—'} icon={Award} href="/theses?status=graduated" />
        <StatCard title="Rata-rata Selesai" value={s?.academic_summary.avg_completion_months ?? '—'} suffix="bln" icon={BookOpen} href="/theses" />
        <StatCard title="Login Hari Ini" value={op?.activity_stats.logins_today ?? '—'} icon={Activity} href="/admin/audit-logs" />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <div>
              <CardTitle>Funnel Tahapan Skripsi</CardTitle>
              <p className="mt-0.5 text-sm text-muted-foreground">Distribusi mahasiswa per tahap</p>
            </div>
            <ViewAllLink href="/theses" />
          </CardHeader>
          <CardContent>
            {summary.isLoading ? (
              <Skeleton className="h-40 w-full" />
            ) : (
              <div className="flex h-40 items-end gap-2">
                {byStatus.map((row) => {
                  const max = Math.max(1, ...byStatus.map((r) => r.count))
                  return (
                    <div key={row.status} className="flex h-full flex-1 flex-col items-center justify-end gap-1.5">
                      <span className="text-xs font-semibold tabular-nums">{row.count}</span>
                      <div
                        className={cn(
                          'w-full rounded-t-sm transition-all',
                          row.status === 'graduated' ? 'bg-success' : 'bg-primary',
                        )}
                        style={{ height: `${Math.max(4, (row.count / max) * 62)}%` }}
                      />
                      <span className="line-clamp-2 text-center text-[11px] leading-tight text-muted-foreground">{row.label}</span>
                    </div>
                  )
                })}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Pending Actions</CardTitle>
            <p className="mt-0.5 text-sm text-muted-foreground">Tindakan yang menunggu Anda</p>
          </CardHeader>
          <CardContent className="space-y-2.5">
            {operational.isLoading && <Skeleton className="h-32 w-full" />}
            {pending && (
              <>
                <div className={ROW}>
                  <span>Review judul</span>
                  <Badge variant="warning">{pending.pending_title_reviews}</Badge>
                </div>
                <div className={ROW}>
                  <span>Review dokumen</span>
                  <Badge variant="warning">{pending.pending_document_reviews}</Badge>
                </div>
                <div className={ROW}>
                  <span>Seminar menunggu</span>
                  <Badge variant="warning">{pending.pending_seminars}</Badge>
                </div>
                <div className={ROW}>
                  <span>Sidang menunggu</span>
                  <Badge variant="warning">{pending.pending_defenses}</Badge>
                </div>
              </>
            )}
            {!operational.isLoading && pending && totalPending === 0 && (
              <p className="py-2 text-sm text-muted-foreground">Tidak ada tindakan tertunda.</p>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <div>
              <CardTitle>Jadwal Mendatang</CardTitle>
              <p className="mt-0.5 text-sm text-muted-foreground">Seminar & sidang 14 hari ke depan</p>
            </div>
            <ViewAllLink href="/schedules" />
          </CardHeader>
          <CardContent className="space-y-2.5">
            {operational.isLoading && <Skeleton className="h-20 w-full" />}
            {schedules.slice(0, 5).map((item) => (
              <div key={`${item.type}-${item.id}`} className="flex items-center gap-3 rounded-md border border-border bg-muted/40 px-4 py-3">
                <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-primary-50 text-primary">
                  <CalendarDays className="h-4 w-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{item.student_name ?? item.thesis_title}</p>
                  <p className="truncate text-xs text-muted-foreground">
                    {item.type === 'seminar' ? 'Seminar' : 'Sidang'} · {formatDateTime(item.scheduled_at)}
                  </p>
                </div>
                {item.room && <Badge variant="muted">{item.room}</Badge>}
              </div>
            ))}
            {!operational.isLoading && schedules.length === 0 && (
              <p className="py-2 text-sm text-muted-foreground">Tidak ada jadwal mendatang.</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Aktivitas Sistem</CardTitle>
            <p className="mt-0.5 text-sm text-muted-foreground">Statistik minggu ini</p>
          </CardHeader>
          <CardContent className="space-y-2.5">
            {operational.isLoading && <Skeleton className="h-24 w-full" />}
            {op?.activity_stats && (
              <>
                <div className={ROW}>
                  <span className="inline-flex items-center gap-2">
                    <ClipboardCheck className="h-4 w-4 text-primary" /> Login hari ini
                  </span>
                  <span className="font-semibold tabular-nums">{op.activity_stats.logins_today}</span>
                </div>
                <div className={ROW}>
                  <span className="inline-flex items-center gap-2">
                    <BookOpen className="h-4 w-4 text-primary" /> Dokumen minggu ini
                  </span>
                  <span className="font-semibold tabular-nums">{op.activity_stats.documents_uploaded_this_week}</span>
                </div>
                <div className={ROW}>
                  <span className="inline-flex items-center gap-2">
                    <GraduationCap className="h-4 w-4 text-primary" /> Bimbingan minggu ini
                  </span>
                  <span className="font-semibold tabular-nums">{op.activity_stats.consultations_this_week}</span>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
