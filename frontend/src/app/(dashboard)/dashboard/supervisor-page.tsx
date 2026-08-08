'use client'

import { useQuery } from '@tanstack/react-query'
import { AlertTriangle, CalendarDays, ClipboardCheck, FilePen, FolderOpen, MessagesSquare } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard, ViewAllLink } from '@/components/features/dashboard/stat-card'
import { DashboardHeader } from '@/components/features/dashboard/dashboard-header'
import { StatusBadge, thesisStatusProps } from '@/components/features/dashboard/status-badge'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { titleChangeApi } from '@/lib/api/title-change-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { flattenSchedules } from '@/types/dashboard'
import { formatDateTime } from '@/lib/utils/date'

export default function SupervisorDashboardPage() {
  const user = useAuthStore((s) => s.user)
  const dash = useQuery({
    queryKey: ['dashboard', 'supervisor'],
    queryFn: dashboardApi.supervisor,
    refetchInterval: 5 * 60 * 1000,
  })

  const d = dash.data
  const schedules = flattenSchedules(d?.upcoming_schedules)

  const pendingTitleChanges = useQuery({
    queryKey: ['title-change', 'pending'],
    queryFn: titleChangeApi.listPending,
  })

  const needsAttention = d?.students?.filter((st) => st.pending_document_reviews > 0) ?? []

  return (
    <div className="space-y-8">
      <DashboardHeader
        name={user?.full_name ?? 'Pembimbing'}
        subtitle={`${roleLabel(user?.role)} · NIDN ${user?.nim_nidn ?? '—'}`}
      />

      {/* Stat cards */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard title="Total Bimbingan" value={d?.total_students ?? '—'} icon={MessagesSquare} href="/supervision" />
        <StatCard title="Dokumen Pending" value={d?.pending_document_reviews ?? '—'} icon={FolderOpen} href="/documents" />
        <StatCard title="Perubahan Judul Pending" value={pendingTitleChanges.data?.length ?? '—'} icon={FilePen} href="/title-change-reviews" />
        <StatCard title="Jadwal 14 Hari" value={schedules.length} icon={CalendarDays} href="/schedules" />
      </div>

      {/* Students + Needs attention */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardContent className="p-5 sm:p-6">
            <div className="mb-4 flex items-start justify-between gap-3">
              <div>
                <h2 className="font-display text-lg leading-none text-st-text">Mahasiswa Bimbingan</h2>
                <p className="mt-0.5 text-[13px] text-st-muted">Daftar mahasiswa yang Anda bimbing</p>
              </div>
              <ViewAllLink href="/supervision" />
            </div>
            <div className="space-y-2">
              {dash.isLoading && <Skeleton className="h-20 w-full" />}
              {d?.students?.map((st) => {
                const { variant, label } = thesisStatusProps(st.status)
                return (
                  <div key={st.thesis_id} className="flex items-center gap-3 rounded-md border border-st-stroke bg-st-surface-hi/60 px-4 py-3">
                    <div className="min-w-0 flex-1">
                      <p className="truncate text-sm font-medium text-st-text">{st.student.full_name}</p>
                      <p className="truncate text-xs text-st-muted">
                        {st.student.nim ?? ''} · {st.consultation_count} bimbingan
                      </p>
                    </div>
                    <div className="flex shrink-0 flex-col items-end gap-1.5">
                      <StatusBadge variant={variant} label={label} />
                      {st.pending_document_reviews > 0 && (
                        <StatusBadge variant="pending" label={`${st.pending_document_reviews} dokumen`} />
                      )}
                    </div>
                  </div>
                )
              })}
              {!dash.isLoading && (!d?.students || d.students.length === 0) && (
                <p className="py-2 text-sm text-st-muted">Belum ada mahasiswa bimbingan.</p>
              )}
            </div>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-5 sm:p-6">
            <h2 className="font-display text-lg leading-none text-st-text">Butuh Perhatian</h2>
            <p className="mt-0.5 text-[13px] text-st-muted">Mahasiswa dengan dokumen pending review</p>
            <div className="mt-4 space-y-2">
              {dash.isLoading && <Skeleton className="h-20 w-full" />}
              {needsAttention.map((st) => (
                <div key={st.thesis_id} className="flex items-center gap-3 rounded-md border border-warning/25 bg-warning-50 px-4 py-3">
                  <AlertTriangle className="h-4 w-4 shrink-0 text-warning" />
                  <div className="min-w-0 flex-1">
                    <p className="truncate text-sm font-medium text-st-text">{st.student.full_name}</p>
                    <p className="text-xs text-st-muted">{st.pending_document_reviews} dokumen menunggu review</p>
                  </div>
                  <ViewAllLink href="/documents" label="Review" />
                </div>
              ))}
              {!dash.isLoading && needsAttention.length === 0 && (
                <p className="py-2 text-sm text-st-muted">Tidak ada yang perlu ditindaklanjuti.</p>
              )}
            </div>
          </CardContent>
        </Card>
      </div>

      {/* Upcoming schedules */}
      <Card>
        <CardContent className="p-5 sm:p-6">
          <div className="mb-4 flex items-start justify-between gap-3">
            <div>
              <h2 className="font-display text-lg leading-none text-st-text">Jadwal Mendatang</h2>
              <p className="mt-0.5 text-[13px] text-st-muted">Seminar & sidang mahasiswa bimbingan</p>
            </div>
            <ViewAllLink href="/schedules" />
          </div>
          <div className="space-y-2">
            {schedules.map((u) => (
              <div key={`${u.type}-${u.id}`} className="flex items-center gap-3 rounded-md border border-st-stroke bg-st-surface-hi/60 px-4 py-3">
                <ClipboardCheck className="h-4 w-4 shrink-0 text-primary" />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium text-st-text">{u.student_name ?? u.thesis_title}</p>
                  <p className="text-xs text-st-muted">
                    {u.type === 'seminar' ? 'Seminar' : 'Sidang'} · {formatDateTime(u.scheduled_at)}
                  </p>
                </div>
                {u.room && (
                  <span className="shrink-0 rounded-md bg-surface-hi px-2 py-0.5 text-xs font-medium text-st-muted">{u.room}</span>
                )}
              </div>
            ))}
            {!dash.isLoading && schedules.length === 0 && (
              <p className="py-2 text-sm text-st-muted">Tidak ada jadwal mendatang.</p>
            )}
          </div>
        </CardContent>
      </Card>
    </div>
  )
}

