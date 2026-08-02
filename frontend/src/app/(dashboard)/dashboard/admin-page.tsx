'use client'

import { useQuery } from '@tanstack/react-query'
import { GraduationCap, Users, BookOpen, Award, ClipboardCheck, CalendarDays, Activity } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard, ViewAllLink } from '@/components/features/dashboard/stat-card'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { cn } from '@/lib/utils/cn'
import { flattenSchedules } from '@/types/dashboard'

function greeting(): string {
  const h = new Date().getHours()
  if (h < 11) return 'Selamat pagi'
  if (h < 15) return 'Selamat siang'
  if (h < 18) return 'Selamat sore'
  return 'Selamat malam'
}

export default function AdminDashboardPage() {
  const { user } = useAuthStore()
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

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">{greeting()}, {user?.full_name ?? 'Admin'} 👋</h1>
        <p className="mt-1 text-sm text-muted-foreground">
          Ringkasan operasional fakultas — {roleLabel(user?.role)}
        </p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard title="Mahasiswa Aktif" value={s?.academic_summary.total_active ?? '—'} icon={Users} iconClass="bg-primary-50 text-primary" href="/admin/users?role=mahasiswa" />
        <StatCard title="Total Lulus" value={s?.academic_summary.total_graduated ?? '—'} icon={Award} iconClass="bg-success-50 text-success" href="/theses?status=graduated" />
        <StatCard title="Rata-rata Selesai" value={s?.academic_summary.avg_completion_months ?? '—'} suffix="bln" icon={BookOpen} iconClass="bg-secondary-50 text-secondary" href="/theses" />
        <StatCard title="Aktivitas Hari Ini" value={op?.activity_stats.logins_today ?? '—'} icon={Activity} iconClass="bg-primary-50 text-primary" href="/admin/audit-logs" />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <div>
              <CardTitle>Funnel Tahapan Skripsi</CardTitle>
              <CardDescription>Distribusi mahasiswa per tahap</CardDescription>
            </div>
            <ViewAllLink href="/theses" />
          </CardHeader>
          <CardContent>
            {summary.isLoading ? (
              <Skeleton className="h-40 w-full" />
            ) : (
              <div className="flex h-44 items-end gap-3">
                {(s?.by_status ?? []).map((row) => {
                  const max = Math.max(1, ...(s?.by_status ?? []).map((r) => r.count))
                  return (
                    <div key={row.status} className="flex flex-1 flex-col items-center gap-1.5">
                      <span className="text-sm font-semibold">{row.count}</span>
                      <div
                        className={cn(
                          'w-full rounded-t-md transition-all',
                          row.status === 'graduated' ? 'bg-success' : 'bg-primary-600',
                        )}
                        style={{ height: `${Math.max(4, (row.count / max) * 100)}%` }}
                      />
                      <span className="text-center text-[10px] leading-tight text-muted-foreground">{row.label}</span>
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
            <CardDescription>Tindakan yang menunggu Anda</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {operational.isLoading && <Skeleton className="h-32 w-full" />}
            {pending && (
              <>
                <div className="flex items-center justify-between rounded-lg border border-warning/30 bg-warning-50 px-3 py-2.5 text-sm">
                  <span>Review judul</span>
                  <Badge variant="warning">{pending.pending_title_reviews}</Badge>
                </div>
                <div className="flex items-center justify-between rounded-lg border border-warning/30 bg-warning-50 px-3 py-2.5 text-sm">
                  <span>Review dokumen</span>
                  <Badge variant="warning">{pending.pending_document_reviews}</Badge>
                </div>
                <div className="flex items-center justify-between rounded-lg border border-warning/30 bg-warning-50 px-3 py-2.5 text-sm">
                  <span>Seminar menunggu</span>
                  <Badge variant="warning">{pending.pending_seminars}</Badge>
                </div>
                <div className="flex items-center justify-between rounded-lg border border-warning/30 bg-warning-50 px-3 py-2.5 text-sm">
                  <span>Sidang menunggu</span>
                  <Badge variant="warning">{pending.pending_defenses}</Badge>
                </div>
              </>
            )}
            {!operational.isLoading && pending && (pending.pending_title_reviews + pending.pending_document_reviews + pending.pending_seminars + pending.pending_defenses) === 0 && (
              <p className="text-sm text-muted-foreground">Tidak ada tindakan tertunda. 🎉</p>
            )}
          </CardContent>
        </Card>
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <div>
              <CardTitle>Jadwal Mendatang</CardTitle>
              <CardDescription>Seminar & sidang 14 hari ke depan</CardDescription>
            </div>
            <ViewAllLink href="/schedules" />
          </CardHeader>
          <CardContent className="space-y-3">
            {operational.isLoading && <Skeleton className="h-24 w-full" />}
            {schedules.slice(0, 5).map((item) => (
              <div key={`${item.type}-${item.id}`} className="flex items-center gap-3 rounded-lg border border-border p-3">
                <div className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary-50 text-primary">
                  <CalendarDays className="h-4 w-4" />
                </div>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{item.student_name ?? item.thesis_title}</p>
                  <p className="truncate text-xs text-muted-foreground">
                    {item.type === 'seminar' ? 'Seminar' : 'Sidang'} · {item.scheduled_at ? new Date(item.scheduled_at).toLocaleString('id-ID') : '—'}
                  </p>
                </div>
                {item.room && <Badge variant="muted">{item.room}</Badge>}
              </div>
            ))}
            {!operational.isLoading && schedules.length === 0 && (
              <p className="text-sm text-muted-foreground">Tidak ada jadwal mendatang.</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Aktivitas Sistem</CardTitle>
            <CardDescription>Statistik minggu ini</CardDescription>
          </CardHeader>
          <CardContent className="space-y-2">
            {operational.isLoading && <Skeleton className="h-24 w-full" />}
            {op?.activity_stats && (
              <>
                <div className="flex items-center justify-between rounded-lg border border-border px-3 py-2.5 text-sm">
                  <span className="inline-flex items-center gap-2"><ClipboardCheck className="h-4 w-4 text-primary" /> Login hari ini</span>
                  <span className="font-semibold">{op.activity_stats.logins_today}</span>
                </div>
                <div className="flex items-center justify-between rounded-lg border border-border px-3 py-2.5 text-sm">
                  <span className="inline-flex items-center gap-2"><BookOpen className="h-4 w-4 text-primary" /> Dokumen minggu ini</span>
                  <span className="font-semibold">{op.activity_stats.documents_uploaded_this_week}</span>
                </div>
                <div className="flex items-center justify-between rounded-lg border border-border px-3 py-2.5 text-sm">
                  <span className="inline-flex items-center gap-2"><GraduationCap className="h-4 w-4 text-primary" /> Bimbingan minggu ini</span>
                  <span className="font-semibold">{op.activity_stats.consultations_this_week}</span>
                </div>
              </>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
