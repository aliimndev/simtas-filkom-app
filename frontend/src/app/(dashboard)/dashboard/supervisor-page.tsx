'use client'

import { useQuery } from '@tanstack/react-query'
import { AlertTriangle, CalendarDays, ClipboardCheck, FilePen, FolderOpen, MessagesSquare } from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard, ViewAllLink } from '@/components/features/dashboard/stat-card'
import { DashboardHeader } from '@/components/features/dashboard/dashboard-header'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { titleChangeApi } from '@/lib/api/title-change-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { flattenSchedules } from '@/types/dashboard'
import { formatDateTime } from '@/lib/utils/date'

const ROW = 'flex items-center gap-3 rounded-md border border-border bg-muted/40 px-4 py-3'

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
    <div className="space-y-6">
      <DashboardHeader
        name={user?.full_name ?? 'Pembimbing'}
        subtitle={`${roleLabel(user?.role)} · NIDN ${user?.nim_nidn ?? '—'}`}
      />

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 xl:grid-cols-4">
        <StatCard title="Total Bimbingan" value={d?.total_students ?? '—'} icon={MessagesSquare} href="/supervision" />
        <StatCard title="Dokumen Pending" value={d?.pending_document_reviews ?? '—'} icon={FolderOpen} href="/documents" />
        <StatCard title="Perubahan Judul Pending" value={pendingTitleChanges.data?.length ?? '—'} icon={FilePen} href="/title-change-reviews" />
        <StatCard title="Jadwal 14 Hari" value={schedules.length} icon={CalendarDays} href="/schedules" />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <div>
              <CardTitle>Mahasiswa Bimbingan</CardTitle>
              <p className="mt-0.5 text-sm text-muted-foreground">Daftar mahasiswa yang Anda bimbing</p>
            </div>
            <ViewAllLink href="/supervision" />
          </CardHeader>
          <CardContent className="space-y-2.5">
            {dash.isLoading && <Skeleton className="h-20 w-full" />}
            {d?.students?.map((st) => (
              <div key={st.thesis_id} className={ROW}>
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{st.student.full_name}</p>
                  <p className="truncate text-xs text-muted-foreground">
                    {st.student.nim ?? ''} · {st.consultation_count} bimbingan
                  </p>
                </div>
                <div className="flex shrink-0 flex-col items-end gap-1">
                  <Badge variant="primary">{st.status.replace(/_/g, ' ')}</Badge>
                  {st.pending_document_reviews > 0 && (
                    <Badge variant="warning">{st.pending_document_reviews} dokumen</Badge>
                  )}
                </div>
              </div>
            ))}
            {!dash.isLoading && (!d?.students || d.students.length === 0) && (
              <p className="py-2 text-sm text-muted-foreground">Belum ada mahasiswa bimbingan.</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader className="flex-row items-center justify-between space-y-0">
            <div>
              <CardTitle>Butuh Perhatian</CardTitle>
              <p className="mt-0.5 text-sm text-muted-foreground">Mahasiswa dengan dokumen pending review</p>
            </div>
          </CardHeader>
          <CardContent className="space-y-2.5">
            {dash.isLoading && <Skeleton className="h-20 w-full" />}
            {needsAttention.map((st) => (
              <div key={st.thesis_id} className="flex items-center gap-3 rounded-md border border-warning/25 bg-warning-50 px-4 py-3">
                <AlertTriangle className="h-4 w-4 shrink-0 text-warning" />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{st.student.full_name}</p>
                  <p className="text-xs text-muted-foreground">{st.pending_document_reviews} dokumen menunggu review</p>
                </div>
                <ViewAllLink href="/documents" label="Review" />
              </div>
            ))}
            {!dash.isLoading && needsAttention.length === 0 && (
              <p className="py-2 text-sm text-muted-foreground">Tidak ada yang perlu ditindaklanjuti.</p>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader className="flex-row items-center justify-between space-y-0">
          <div>
            <CardTitle>Jadwal Mendatang</CardTitle>
            <p className="mt-0.5 text-sm text-muted-foreground">Seminar & sidang mahasiswa bimbingan</p>
          </div>
          <ViewAllLink href="/schedules" />
        </CardHeader>
        <CardContent className="space-y-2.5">
          {schedules.map((u) => (
            <div key={`${u.type}-${u.id}`} className={ROW}>
              <ClipboardCheck className="h-4 w-4 shrink-0 text-primary" />
              <div className="min-w-0 flex-1">
                <p className="truncate text-sm font-medium">{u.student_name ?? u.thesis_title}</p>
                <p className="text-xs text-muted-foreground">
                  {u.type === 'seminar' ? 'Seminar' : 'Sidang'} · {formatDateTime(u.scheduled_at)}
                </p>
              </div>
              {u.room && <Badge variant="muted">{u.room}</Badge>}
            </div>
          ))}
          {!dash.isLoading && schedules.length === 0 && (
            <p className="py-2 text-sm text-muted-foreground">Tidak ada jadwal mendatang.</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
