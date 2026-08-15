'use client'

import { useQuery } from '@tanstack/react-query'
import { Activity, Award, BookOpen, CalendarDays, ClipboardCheck, GraduationCap, Users } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard, ViewAllLink } from '@/components/features/dashboard/stat-card'
import { DashboardHeader } from '@/components/features/dashboard/dashboard-header'
import { DashboardHeroBand } from '@/components/features/dashboard/dashboard-hero-band'
import { StatusBadge } from '@/components/features/dashboard/status-badge'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { cn } from '@/lib/utils/cn'
import { flattenSchedules } from '@/types/dashboard'
import { formatDateTime } from '@/lib/utils/date'

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

  const pendingRows = pending
    ? [
        { label: 'Review judul', count: pending.pending_title_reviews, href: '/theses' },
        { label: 'Review dokumen', count: pending.pending_document_reviews, href: '/documents' },
        { label: 'Seminar menunggu', count: pending.pending_seminars, href: '/seminars' },
        { label: 'Sidang menunggu', count: pending.pending_defenses, href: '/defenses' },
      ]
    : []

  return (
    <div className="space-y-8">
      <DashboardHeroBand>
        <DashboardHeader
          name={user?.full_name ?? 'Admin'}
          subtitle={`Ringkasan operasional fakultas · ${roleLabel(user?.role)}`}
        />
      </DashboardHeroBand>

      {/* Stat cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard title="Mahasiswa Aktif" value={s?.academic_summary.total_active ?? '—'} icon={Users} href="/admin/users?role=mahasiswa" />
        <StatCard title="Total Lulus" value={s?.academic_summary.total_graduated ?? '—'} icon={Award} href="/theses?status=graduated" />
        <StatCard title="Rata-rata Selesai" value={s?.academic_summary.avg_completion_months ?? '—'} suffix="bln" icon={BookOpen} href="/theses" />
        <StatCard title="Login Hari Ini" value={op?.activity_stats.logins_today ?? '—'} icon={Activity} href="/admin/audit-logs" />
      </div>

      {/* Funnel + Pending */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardContent className="p-5 sm:p-6">
            <div className="mb-5 flex items-start justify-between gap-3">
              <div>
                <h2 className="landing-heading text-lg">Funnel Tahapan <span className="accent-text italic">Skripsi</span></h2>
                <p className="mt-0.5 text-[13px] text-st-muted">Distribusi mahasiswa per tahap</p>
              </div>
              <ViewAllLink href="/theses" />
            </div>
            {summary.isLoading ? (
              <Skeleton className="h-40 w-full" />
            ) : (
              <div className="flex h-40 items-end gap-2">
                {byStatus.map((row) => {
                  const max = Math.max(1, ...byStatus.map((r) => r.count))
                  return (
                    <div key={row.status} className="flex h-full flex-1 flex-col items-center justify-end gap-1.5">
                      <span className="text-xs font-semibold tabular-nums text-st-text">{row.count}</span>
                      <div
                        className={cn(
                          'w-full rounded-t-sm transition-all',
                          row.status === 'graduated' ? 'bg-success' : 'bg-primary',
                        )}
                        style={{ height: `${Math.max(4, (row.count / max) * 62)}%` }}
                      />
                      <span className="line-clamp-2 text-center text-[11px] leading-tight text-st-muted">{row.label}</span>
                    </div>
                  )
                })}
              </div>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-5 sm:p-6">
            <h2 className="landing-heading text-lg">Pending <span className="accent-text italic">Actions</span></h2>
            <p className="mt-0.5 text-[13px] text-st-muted">Tindakan yang menunggu Anda</p>
            <div className="mt-4 space-y-2">
              {operational.isLoading && <Skeleton className="h-32 w-full" />}
              {pendingRows.map((row) => (
                <div key={row.label} className="flex items-center justify-between gap-3 rounded-xl border border-st-stroke bg-st-surface-hi/60 px-4 py-3">
                  <span className="text-sm text-st-text">{row.label}</span>
                  {row.count > 0 ? (
                    <StatusBadge variant="pending" label={`${row.count} menunggu`} />
                  ) : (
                    <span className="text-xs font-medium text-st-muted">Selesai</span>
                  )}
                </div>
              ))}
              {!operational.isLoading && totalPending === 0 && (
                <p className="py-2 text-sm text-st-muted">Tidak ada tindakan tertunda.</p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>


      {/* Upcoming + System activity */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardContent className="p-5 sm:p-6">
            <div className="mb-4 flex items-start justify-between gap-3">
              <div>
                <h2 className="landing-heading text-lg">Jadwal <span className="accent-text italic">Mendatang</span></h2>
                <p className="mt-0.5 text-[13px] text-st-muted">Seminar & sidang 14 hari ke depan</p>
              </div>
              <ViewAllLink href="/schedules" />
            </div>
            <div className="space-y-2">
              {operational.isLoading && <Skeleton className="h-20 w-full" />}
              {schedules.slice(0, 5).map((item) => (
                <div key={`${item.type}-${item.id}`} className="flex items-center gap-3 rounded-xl border border-st-stroke bg-st-surface-hi/60 px-4 py-3">
                  <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-(--st-accent-from)/10 text-(--st-accent-to)">
                    <CalendarDays className="h-4 w-4" />
                  </div>
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-st-text">{item.student_name ?? item.thesis_title}</p>
                    <p className="truncate text-xs text-st-muted">
                      {item.type === 'seminar' ? 'Seminar' : 'Sidang'} · {formatDateTime(item.scheduled_at)}
                    </p>
                  </div>
                  {item.room && (
                    <span className="shrink-0 rounded-full bg-st-surface-hi px-2 py-0.5 text-xs font-medium text-st-muted">{item.room}</span>
                  )}
                </div>
              ))}
              {!operational.isLoading && schedules.length === 0 && (
                <p className="py-2 text-sm text-st-muted">Tidak ada jadwal mendatang.</p>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-5 sm:p-6">
            <h2 className="landing-heading text-lg">Aktivitas <span className="accent-text italic">Sistem</span></h2>
            <p className="mt-0.5 text-[13px] text-st-muted">Statistik minggu ini</p>
            <div className="mt-4 space-y-2">
              {operational.isLoading && <Skeleton className="h-24 w-full" />}
              {op?.activity_stats && (
                <>
                  <div className="flex items-center justify-between gap-3 rounded-xl border border-st-stroke bg-st-surface-hi/60 px-4 py-3">
                    <span className="inline-flex items-center gap-2 text-sm text-st-text">
                      <ClipboardCheck className="h-4 w-4 text-(--st-accent-to)" /> Login hari ini
                    </span>
                    <span className="font-semibold tabular-nums text-st-text">{op.activity_stats.logins_today}</span>
                  </div>
                  <div className="flex items-center justify-between gap-3 rounded-xl border border-st-stroke bg-st-surface-hi/60 px-4 py-3">
                    <span className="inline-flex items-center gap-2 text-sm text-st-text">
                      <BookOpen className="h-4 w-4 text-(--st-accent-to)" /> Dokumen minggu ini
                    </span>
                    <span className="font-semibold tabular-nums text-st-text">{op.activity_stats.documents_uploaded_this_week}</span>
                  </div>
                  <div className="flex items-center justify-between gap-3 rounded-xl border border-st-stroke bg-st-surface-hi/60 px-4 py-3">
                    <span className="inline-flex items-center gap-2 text-sm text-st-text">
                      <GraduationCap className="h-4 w-4 text-(--st-accent-to)" /> Bimbingan minggu ini
                    </span>
                    <span className="font-semibold tabular-nums text-st-text">{op.activity_stats.consultations_this_week}</span>
                  </div>
                </>
              )}
            </div>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

