'use client'

import { useQuery } from '@tanstack/react-query'
import { AxiosError } from 'axios'
import {
  AlertCircle,
  ArrowRight,
  BookOpen,
  CalendarDays,
  FileText,
  FolderOpen,
  MessagesSquare,
  Upload,
  PenLine,
} from 'lucide-react'
import Link from 'next/link'
import { Card, CardContent } from '@/components/ui/card'
import { Skeleton } from '@/components/ui/skeleton'
import { DashboardHeader } from '@/components/features/dashboard/dashboard-header'
import { ProgressTimeline } from '@/components/features/dashboard/progress-timeline'
import { StatusBadge, thesisStatusProps } from '@/components/features/dashboard/status-badge'
import { ActivityList, type ActivityItem } from '@/components/features/dashboard/activity-list'
import { UpcomingList, type UpcomingItem } from '@/components/features/dashboard/upcoming-list'
import { QuickActions, type QuickActionItem } from '@/components/features/dashboard/quick-actions'
import { StatCard } from '@/components/features/dashboard/stat-card'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { formatDate, formatDateTime } from '@/lib/utils/date'

export default function StudentDashboardPage() {
  const user = useAuthStore((s) => s.user)
  const dash = useQuery({
    queryKey: ['dashboard', 'student'],
    queryFn: dashboardApi.student,
    refetchInterval: 5 * 60 * 1000,
    retry: false,
  })

  const d = dash.data
  const isNoThesis =
    !d && dash.error instanceof AxiosError && dash.error.response?.status === 404

  if (dash.isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-72" />
        <Skeleton className="h-4 w-96" />
        <Skeleton className="h-48 w-full" />
        <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
          <Skeleton className="h-24 w-full" />
        </div>
      </div>
    )
  }

  if (dash.isError && !isNoThesis) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <AlertCircle className="mx-auto h-9 w-9 text-danger" />
          <p className="mt-3 font-semibold">Gagal memuat dashboard</p>
          <p className="mt-1 text-sm text-muted-foreground">Silakan muat ulang halaman atau coba lagi nanti.</p>
        </CardContent>
      </Card>
    )
  }

  if (isNoThesis || !d) {
    return (
      <div className="space-y-8">
        <DashboardHeader
          name={user?.full_name ?? 'Mahasiswa'}
          subtitle={`${roleLabel(user?.role)} · NIM ${user?.nim_nidn ?? '—'}`}
        />
        <Card className="border-dashed border-st-stroke bg-st-surface">
          <CardContent className="flex flex-col items-center gap-5 px-6 py-16 text-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-full bg-primary-50 text-primary">
              <BookOpen className="h-6 w-6" />
            </div>
            <div className="max-w-md">
              <h2 className="font-display text-2xl leading-tight text-st-text">Mulai Ajukan Judul Skripsi</h2>
              <p className="mt-2 text-sm text-st-muted">
                Ajukan judul dan abstrak skripsi Anda untuk direview oleh Kaprodi sebelum memulai bimbingan.
              </p>
            </div>
            <Link
              href="/thesis/new"
              className="accent-ring inline-flex h-11 items-center gap-2 rounded-full bg-st-text px-7 text-sm font-medium text-st-bg transition hover:opacity-90"
            >
              Ajukan Judul Skripsi <ArrowRight className="h-4 w-4" />
            </Link>
          </CardContent>
        </Card>
      </div>
    )
  }

  const totalDocs = (d.documents ?? []).length
  const pendingActions = d.pending_actions ?? []
  const { variant: statusVariant, label: statusLabel } = thesisStatusProps(d.status)

  // Build activity items from available data
  const activityItems: ActivityItem[] = buildActivityItems(d)

  // Build upcoming items from schedule data
  const upcomingItems: UpcomingItem[] = buildUpcomingItems(d)

  // Quick actions
  const quickActions: QuickActionItem[] = [
    { label: 'Upload Dokumen', href: '/documents', icon: Upload, variant: 'secondary' },
    { label: 'Buat Catatan Bimbingan', href: '/supervision', icon: PenLine, variant: 'secondary' },
  ]

  return (
    <div className="space-y-8">
      {/* Header */}
      <div>
        <DashboardHeader
          name={user?.full_name ?? 'Mahasiswa'}
          subtitle="Pantau seluruh proses Tugas Akhir Skripsi Anda dalam satu tempat."
        />
      </div>

      {/* Thesis Status Card — prominent section */}
      <Card className="border-st-stroke bg-st-surface">
        <CardContent className="p-5 sm:p-6">
          <div className="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2.5">
                <p className="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-st-muted">Tugas Akhir Skripsi</p>
                <StatusBadge variant={statusVariant} label={statusLabel} />
              </div>
              <h2 className="mt-2.5 font-display text-2xl leading-tight text-st-text sm:text-3xl">
                {d.title}
              </h2>
              <dl className="mt-3 flex flex-wrap gap-x-6 gap-y-1.5 text-sm text-st-muted">
                <div className="flex items-center gap-1.5">
                  <dt className="sr-only">Pembimbing</dt>
                  <dd>{d.supervisors?.length ? d.supervisors.map((s) => s.full_name).join(', ') : 'Belum ditentukan'}</dd>
                </div>
                <div className="flex items-center gap-1.5">
                  <dt className="sr-only">Progres</dt>
                  <dd>{d.progress_percentage}% selesai</dd>
                </div>
                {d.last_consultation && (
                  <div className="flex items-center gap-1.5">
                    <dt className="sr-only">Terakhir diperbarui</dt>
                    <dd>Diperbarui {formatDate(d.last_consultation)}</dd>
                  </div>
                )}
              </dl>
            </div>
          </div>
          {/* Progress bar (compact) */}
          <div className="mt-5">
            <div className="h-1.5 w-full overflow-hidden rounded-full bg-st-surface-hi">
              <div
                className="accent-gradient h-full rounded-full transition-all duration-500"
                style={{ width: `${d.progress_percentage}%` }}
              />
            </div>
          </div>
        </CardContent>
      </Card>

      {/* Pending Actions */}
      {pendingActions.length > 0 && (
        <Card className="border-warning/30 bg-warning-50/50">
          <CardContent className="p-5 sm:p-6">
            <p className="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-warning">
              <AlertCircle className="mr-1.5 -mt-0.5 inline h-4 w-4" /> Yang perlu Anda lakukan
            </p>
            <ul className="mt-3 space-y-2 text-sm text-st-text">
              {pendingActions.map((a, i) => (
                <li key={i} className="flex items-baseline gap-2">
                  <span className="mt-1.5 h-1 w-1 shrink-0 rounded-full bg-warning" />
                  {a}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

      {/* Progress Timeline */}
      <Card className="border-st-stroke bg-st-surface">
        <CardContent className="p-5 sm:p-6">
          <ProgressTimeline status={d.status} />
        </CardContent>
      </Card>

      {/* Stats row */}
      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard title="Bimbingan" value={d.consultation_count ?? 0} icon={MessagesSquare} href="/supervision" />
        <StatCard title="Dokumen" value={totalDocs} icon={FolderOpen} href="/documents" />
        <StatCard
          title="Ujian Mendatang"
          value={(d.upcoming_seminar ? 1 : 0) + (d.upcoming_defense ? 1 : 0)}
          icon={CalendarDays}
          href="/schedules"
        />
      </div>

      {/* Two-column: Activity + Upcoming & Quick Actions */}
      <div className="grid grid-cols-1 gap-6 lg:grid-cols-5">
        {/* Activity */}
        <Card className="border-st-stroke bg-st-surface lg:col-span-3">
          <CardContent className="p-5 sm:p-6">
            <h2 className="mb-4 font-display text-lg leading-none text-st-text">Aktivitas Terbaru</h2>
            <ActivityList items={activityItems} />
          </CardContent>
        </Card>

        {/* Upcoming + Quick Actions */}
        <div className="flex flex-col gap-6 lg:col-span-2">
          <Card className="border-st-stroke bg-st-surface">
            <CardContent className="p-5 sm:p-6">
              <h2 className="mb-4 font-display text-lg leading-none text-st-text">Agenda Berikutnya</h2>
              <UpcomingList items={upcomingItems} />
            </CardContent>
          </Card>

          <Card className="border-st-stroke bg-st-surface">
            <CardContent className="p-5 sm:p-6">
              <h2 className="mb-4 font-display text-lg leading-none text-st-text">Tindakan Cepat</h2>
              <QuickActions actions={quickActions} />
            </CardContent>
          </Card>
        </div>
      </div>
    </div>
  )
}

/** Build activity items from student dashboard data. */
function buildActivityItems(d: NonNullable<ReturnType<typeof dashboardApi.student> extends Promise<infer T> ? T : never>): ActivityItem[] {
  const items: ActivityItem[] = []

  // Status-based activity
  if (d.status === 'submitted') {
    items.push({
      id: 'status-submitted',
      description: 'Judul skripsi diajukan untuk review',
      type: 'submission',
      timestamp: 'Baru saja',
    })
  }

  // Consultation activity
  if (d.last_consultation) {
    items.push({
      id: 'last-consultation',
      description: 'Sesi bimbingan terakhir tercatat',
      actor: d.supervisors?.[0]?.full_name,
      type: 'note',
      timestamp: formatDateTime(d.last_consultation),
    })
  }

  // Document activities
  ;(d.documents ?? []).slice(0, 2).forEach((doc, i) => {
    items.push({
      id: `doc-${i}`,
      description: `Dokumen ${doc.type} (v${doc.version}) diunggah`,
      type: 'upload',
      timestamp: 'Terbaru',
    })
  })

  // If there are upcoming schedules
  if (d.upcoming_seminar) {
    items.push({
      id: 'seminar-scheduled',
      description: 'Jadwal seminar proposal ditentukan',
      type: 'schedule',
      timestamp: formatDate(d.upcoming_seminar.scheduled_at),
    })
  }

  // Provide at least one item if empty
  if (items.length === 0) {
    items.push({
      id: 'no-activity',
      icon: FileText,
      description: 'Belum ada aktivitas tercatat',
      timestamp: '—',
    })
  }

  return items.slice(0, 6)
}

/** Build upcoming items from student dashboard schedule data. */
function buildUpcomingItems(d: NonNullable<ReturnType<typeof dashboardApi.student> extends Promise<infer T> ? T : never>): UpcomingItem[] {
  const items: UpcomingItem[] = []

  if (d.upcoming_seminar?.scheduled_at) {
    items.push({
      id: 'seminar',
      title: 'Seminar Proposal',
      date: formatDate(d.upcoming_seminar.scheduled_at),
      time: d.upcoming_seminar.room ?? undefined,
      type: 'seminar',
    })
  }

  if (d.upcoming_defense?.scheduled_at) {
    items.push({
      id: 'defense',
      title: 'Sidang Akhir',
      date: formatDate(d.upcoming_defense.scheduled_at),
      time: d.upcoming_defense.room ?? undefined,
      type: 'sidang',
    })
  }

  return items
}
