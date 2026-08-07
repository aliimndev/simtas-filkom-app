'use client'

import { useQuery } from '@tanstack/react-query'
import { AxiosError } from 'axios'
import { AlertCircle, ArrowRight, BookOpen, CalendarDays, FolderOpen, MessagesSquare } from 'lucide-react'
import Link from 'next/link'
import { Card, CardContent } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard } from '@/components/features/dashboard/stat-card'
import { DashboardHeader } from '@/components/features/dashboard/dashboard-header'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { formatDate } from '@/lib/utils/date'
import { cn } from '@/lib/utils/cn'

const STAGE_ORDER = ['submitted', 'approved', 'in_progress', 'seminar_ready', 'seminar_done', 'defense_ready', 'defense_done', 'graduated']
const STAGE_LABELS: Record<string, string> = {
  submitted: 'Pengajuan',
  approved: 'Disetujui',
  in_progress: 'Bimbingan',
  seminar_ready: 'Seminar',
  seminar_done: 'Pasca Seminar',
  defense_ready: 'Sidang',
  defense_done: 'Pasca Sidang',
  graduated: 'Lulus',
}

export default function StudentDashboardPage() {
  const { user } = useAuthStore()
  const dash = useQuery({
    queryKey: ['dashboard', 'student'],
    queryFn: dashboardApi.student,
    refetchInterval: 5 * 60 * 1000,
    retry: false,
  })

  const d = dash.data
  const stageIndex = d ? STAGE_ORDER.indexOf(d.status) : -1
  const isNoThesis =
    !d && dash.error instanceof AxiosError && dash.error.response?.status === 404

  if (dash.isLoading) {
    return (
      <div className="space-y-6">
        <Skeleton className="h-8 w-72" />
        <Skeleton className="h-32 w-full" />
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
      <div className="space-y-6">
        <DashboardHeader
          name={user?.full_name ?? 'Mahasiswa'}
          subtitle={`${roleLabel(user?.role)} · NIM ${user?.nim_nidn ?? '—'}`}
        />
        <Card className="border-dashed">
          <CardContent className="flex flex-col items-center gap-4 px-6 py-14 text-center">
            <div className="flex h-12 w-12 items-center justify-center rounded-md bg-primary-50 text-primary">
              <BookOpen className="h-6 w-6" />
            </div>
            <div className="max-w-md">
              <h2 className="text-lg font-semibold tracking-tight">Mulai Ajukan Judul Skripsi</h2>
              <p className="mt-1 text-sm text-muted-foreground">
                Ajukan judul dan abstrak skripsi Anda untuk direview oleh Kaprodi sebelum memulai bimbingan.
              </p>
            </div>
            <Button asChild>
              <Link href="/thesis/new" className="inline-flex items-center gap-2">
                Ajukan Judul Skripsi <ArrowRight className="h-4 w-4" />
              </Link>
            </Button>
          </CardContent>
        </Card>
      </div>
    )
  }

  // Backend always sends `[]`, but guard anyway so a `null` payload (older
  // backend, or empty collections serialized as null) never crashes the page.
  const totalDocs = (d.documents ?? []).length
  const pendingActions = d.pending_actions ?? []
  const stageLabel = STAGE_LABELS[d.status] ?? d.status.replace(/_/g, ' ')
  const statusVariant = d.status === 'cancelled' ? 'danger' : d.status === 'graduated' ? 'success' : 'primary'

  return (
    <div className="space-y-6">
      <DashboardHeader
        name={user?.full_name ?? 'Mahasiswa'}
        subtitle={`${roleLabel(user?.role)} · NIM ${user?.nim_nidn ?? '—'}`}
      />

      {/* Status & progress */}
      <Card>
        <CardContent className="p-6">
          <div className="flex flex-wrap items-center justify-between gap-3">
            <div>
              <h2 className="text-sm font-semibold">Status skripsi</h2>
              <p className="mt-0.5 text-sm text-muted-foreground">
                Tahap {stageLabel} · progres {d.progress_percentage}%
              </p>
            </div>
            <Badge variant={statusVariant}>{stageLabel}</Badge>
          </div>
          {/* Mobile: horizontal snap scroll | Desktop: flex row */}
          <div className="-mx-1 mt-6 overflow-x-auto pb-1 sm:mx-0 sm:overflow-visible sm:pb-0">
            <div className="flex min-w-[640px] items-center sm:min-w-0">
              {STAGE_ORDER.map((stage, i) => {
                const done = i <= stageIndex
                return (
                  <div key={stage} className="flex flex-1 items-center last:flex-none">
                    <div className="flex flex-col items-center gap-1.5">
                      <div
                        className={cn(
                          'flex h-7 w-7 shrink-0 items-center justify-center rounded-full text-xs font-semibold',
                          done ? 'bg-primary text-primary-foreground' : 'bg-surface-hi text-muted-foreground',
                        )}
                      >
                        {i + 1}
                      </div>
                      <span className="whitespace-nowrap text-[11px] text-muted-foreground">{STAGE_LABELS[stage]}</span>
                    </div>
                    {i < STAGE_ORDER.length - 1 && (
                      <div className={cn('mx-2 h-0.5 flex-1 rounded-full', done ? 'bg-primary' : 'bg-border')} />
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        </CardContent>
      </Card>

      {pendingActions.length > 0 && (
        <Card className="border-warning/30 bg-warning-50/70">
          <CardContent className="p-4">
            <p className="flex items-center gap-2 text-sm font-semibold text-warning-700">
              <AlertCircle className="h-4 w-4" /> Yang perlu Anda lakukan
            </p>
            <ul className="mt-2 space-y-1.5 text-sm text-warning-900">
              {pendingActions.map((a, i) => (
                <li key={i} className="flex items-baseline gap-2">
                  <span className="h-1 w-1 shrink-0 translate-y-[-2px] rounded-full bg-warning" />
                  {a}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

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

      <div className="grid grid-cols-1 gap-6 lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardContent className="p-6">
            <h2 className="text-sm font-semibold">Skripsi Saya</h2>
            <dl className="mt-2 divide-y divide-border">
              <div className="flex flex-col gap-1 py-3 sm:flex-row sm:items-baseline sm:justify-between sm:gap-6">
                <dt className="shrink-0 text-sm text-muted-foreground">Judul</dt>
                <dd className="text-sm font-medium text-foreground">{d.title}</dd>
              </div>
              <div className="flex flex-col gap-1 py-3 sm:flex-row sm:items-baseline sm:justify-between sm:gap-6">
                <dt className="shrink-0 text-sm text-muted-foreground">Pembimbing</dt>
                <dd className="text-sm font-medium text-foreground">
                  {d.supervisors?.length ? d.supervisors.map((s) => s.full_name).join(', ') : 'Belum ditentukan'}
                </dd>
              </div>
              <div className="flex flex-col gap-1 py-3 sm:flex-row sm:items-baseline sm:justify-between sm:gap-6">
                <dt className="shrink-0 text-sm text-muted-foreground">Bimbingan terakhir</dt>
                <dd className="text-sm font-medium text-foreground">
                  {d.last_consultation ? formatDate(d.last_consultation) : 'Belum ada'}
                </dd>
              </div>
            </dl>
          </CardContent>
        </Card>

        <Card>
          <CardContent className="p-6">
            <h2 className="text-sm font-semibold">Jadwal</h2>
            <dl className="mt-2 space-y-3">
              <div className="rounded-md border border-border bg-muted/40 p-3">
                <dt className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Seminar</dt>
                <dd className="mt-0.5 text-sm font-medium">
                  {d.upcoming_seminar ? (
                    <>
                      {d.upcoming_seminar.scheduled_at ? formatDate(d.upcoming_seminar.scheduled_at) : '—'}
                      {d.upcoming_seminar.room ? ` · ${d.upcoming_seminar.room}` : ''}
                    </>
                  ) : (
                    <span className="font-normal text-muted-foreground">Belum dijadwalkan</span>
                  )}
                </dd>
              </div>
              <div className="rounded-md border border-border bg-muted/40 p-3">
                <dt className="text-[11px] font-medium uppercase tracking-wide text-muted-foreground">Sidang</dt>
                <dd className="mt-0.5 text-sm font-medium">
                  {d.upcoming_defense ? (
                    <>
                      {d.upcoming_defense.scheduled_at ? formatDate(d.upcoming_defense.scheduled_at) : '—'}
                      {d.upcoming_defense.room ? ` · ${d.upcoming_defense.room}` : ''}
                    </>
                  ) : (
                    <span className="font-normal text-muted-foreground">Belum dijadwalkan</span>
                  )}
                </dd>
              </div>
            </dl>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
