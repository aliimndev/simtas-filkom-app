'use client'

import { useQuery } from '@tanstack/react-query'
import { AxiosError } from 'axios'
import { AlertCircle, ArrowRight, BookOpen, CalendarDays, FolderOpen, MessagesSquare } from 'lucide-react'
import Link from 'next/link'
import { Card, CardContent } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Skeleton } from '@/components/ui/skeleton'
import { StatCard } from '@/components/features/dashboard/stat-card'
import { DashboardHeader } from '@/components/features/dashboard/dashboard-header'
import { StageStepper } from '@/components/features/dashboard/stage-stepper'
import { PendingActionsCard } from '@/components/features/dashboard/pending-actions-card'
import { ThesisInfoCard } from '@/components/features/dashboard/thesis-info-card'
import { ScheduleCard } from '@/components/features/dashboard/schedule-card'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'

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

  return (
    <div className="space-y-6">
      <DashboardHeader
        name={user?.full_name ?? 'Mahasiswa'}
        subtitle={`${roleLabel(user?.role)} · NIM ${user?.nim_nidn ?? '—'}`}
      />

      <StageStepper status={d.status} progressPercentage={d.progress_percentage} />

      <PendingActionsCard actions={pendingActions} />

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
        <ThesisInfoCard thesis={d} />
        <ScheduleCard seminar={d.upcoming_seminar} defense={d.upcoming_defense} />
      </div>
    </div>
  )
}
