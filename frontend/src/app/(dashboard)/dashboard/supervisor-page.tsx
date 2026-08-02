'use client'

import { useQuery } from '@tanstack/react-query'
import { MessagesSquare, FolderOpen, CalendarDays, ClipboardCheck, AlertTriangle } from 'lucide-react'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard, ViewAllLink } from '@/components/features/dashboard/stat-card'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { flattenSchedules } from '@/types/dashboard'
import { formatDateTime } from '@/lib/utils/date'

function greeting(): string {
  const h = new Date().getHours()
  if (h < 11) return 'Selamat pagi'
  if (h < 15) return 'Selamat siang'
  if (h < 18) return 'Selamat sore'
  return 'Selamat malam'
}

export default function SupervisorDashboardPage() {
  const { user } = useAuthStore()
  const dash = useQuery({
    queryKey: ['dashboard', 'supervisor'],
    queryFn: dashboardApi.supervisor,
    refetchInterval: 5 * 60 * 1000,
  })

  const d = dash.data
  const schedules = flattenSchedules(d?.upcoming_schedules)

  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-2xl font-bold">{greeting()}, {user?.full_name?.split(' ')[0]} 👋</h1>
        <p className="mt-1 text-sm text-muted-foreground">{roleLabel(user?.role)} — NIDN {user?.nim_nidn ?? '—'}</p>
      </div>

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard title="Total Bimbingan" value={d?.total_students ?? '—'} icon={MessagesSquare} href="/supervision" iconClass="bg-primary-50 text-primary" />
        <StatCard title="Dokumen Pending" value={d?.pending_document_reviews ?? '—'} icon={FolderOpen} href="/documents" iconClass="bg-warning-50 text-warning" />
        <StatCard title="Jadwal 14 Hari" value={schedules.length} icon={CalendarDays} href="/schedules" iconClass="bg-success-50 text-success" />
      </div>

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-2">
        <Card>
          <CardHeader>
            <CardTitle>Mahasiswa Bimbingan</CardTitle>
            <CardDescription>Daftar mahasiswa yang Anda bimbing</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {dash.isLoading && <Skeleton className="h-24 w-full" />}
            {d?.students?.map((st) => (
              <div key={st.thesis_id} className="flex items-center justify-between gap-3 rounded-lg border border-border p-3">
                <div className="min-w-0">
                  <p className="truncate text-sm font-medium">{st.student.full_name}</p>
                  <p className="truncate text-xs text-muted-foreground">
                    {st.student.nim ?? ''} · {st.consultation_count} bimbingan
                  </p>
                </div>
                <div className="flex flex-col items-end gap-1">
                  <Badge variant="primary">{st.status.replace(/_/g, ' ')}</Badge>
                  {st.pending_document_reviews > 0 && (
                    <Badge variant="warning">{st.pending_document_reviews} dokumen</Badge>
                  )}
                </div>
              </div>
            ))}
            {!dash.isLoading && (!d?.students || d.students.length === 0) && (
              <p className="text-sm text-muted-foreground">Belum ada mahasiswa bimbingan.</p>
            )}
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Butuh Perhatian</CardTitle>
            <CardDescription>Mahasiswa dengan dokumen pending review</CardDescription>
          </CardHeader>
          <CardContent className="space-y-3">
            {dash.isLoading && <Skeleton className="h-24 w-full" />}
            {d?.students?.filter((st) => st.pending_document_reviews > 0).map((st) => (
              <div key={st.thesis_id} className="flex items-center gap-3 rounded-lg border border-warning/30 bg-warning-50 p-3">
                <AlertTriangle className="h-4 w-4 shrink-0 text-warning" />
                <div className="min-w-0 flex-1">
                  <p className="truncate text-sm font-medium">{st.student.full_name}</p>
                  <p className="text-xs text-muted-foreground">{st.pending_document_reviews} dokumen menunggu review</p>
                </div>
                <ViewAllLink href="/documents" label="Review" />
              </div>
            ))}
            {!dash.isLoading && (!d?.students || d.students.every((st) => st.pending_document_reviews === 0)) && (
              <p className="text-sm text-muted-foreground">Tidak ada yang perlu ditindaklanjuti. 🎉</p>
            )}
          </CardContent>
        </Card>
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Jadwal Mendatang</CardTitle>
          <CardDescription>Seminar & sidang mahasiswa bimbingan</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          {schedules.map((u) => (
            <div key={`${u.type}-${u.id}`} className="flex items-center gap-3 rounded-lg border border-border p-3">
              <ClipboardCheck className="h-4 w-4 text-primary" />
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
            <p className="text-sm text-muted-foreground">Tidak ada jadwal mendatang.</p>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
