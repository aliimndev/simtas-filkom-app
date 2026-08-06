'use client'

import { useQuery } from '@tanstack/react-query'
import { AxiosError } from 'axios'
import { BookOpen, MessagesSquare, FolderOpen, CalendarDays, ArrowRight, AlertCircle } from 'lucide-react'
import Link from 'next/link'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { StatCard } from '@/components/features/dashboard/stat-card'
import { dashboardApi } from '@/lib/api/dashboard-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { formatDate } from '@/lib/utils/date'

function greeting(): string {
  const h = new Date().getHours()
  if (h < 11) return 'Selamat pagi'
  if (h < 15) return 'Selamat siang'
  if (h < 18) return 'Selamat sore'
  return 'Selamat malam'
}

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
      <div className="space-y-4">
        <div className="h-8 w-64 animate-pulse rounded bg-muted" />
        <div className="h-32 animate-pulse rounded-xl bg-muted" />
      </div>
    )
  }

  if (dash.isError && !isNoThesis) {
    return (
      <Card>
        <CardContent className="py-12 text-center">
          <AlertCircle className="mx-auto h-10 w-10 text-danger" />
          <p className="mt-3 font-semibold">Gagal memuat dashboard</p>
          <p className="mt-1 text-sm text-muted-foreground">Silakan muat ulang halaman atau coba lagi nanti.</p>
        </CardContent>
      </Card>
    )
  }

  if (isNoThesis || !d) {
    return (
      <div className="space-y-6">
        <h1 className="text-2xl font-bold">{greeting()}, {user?.full_name?.split(' ')[0]} 👋</h1>
        <Card className="border-2 border-dashed border-primary-200 bg-primary-50/50">
          <CardContent className="flex flex-col items-center gap-4 py-14 text-center">
            <div className="flex h-16 w-16 items-center justify-center rounded-2xl bg-primary text-primary-foreground">
              <BookOpen className="h-8 w-8" />
            </div>
            <div>
              <h2 className="text-xl font-bold">Mulai Ajukan Judul Skripsi Anda</h2>
              <p className="mt-1 max-w-md text-sm text-muted-foreground">
                Ajukan judul dan abstrak skripsi Anda untuk direview oleh Kaprodi sebelum memulai bimbingan.
              </p>
            </div>
            <Button asChild size="lg">
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
      <div>
        <h1 className="text-2xl font-bold">{greeting()}, {user?.full_name?.split(' ')[0]} 👋</h1>
        <p className="mt-1 text-sm text-muted-foreground">{roleLabel(user?.role)} — NIM {user?.nim_nidn ?? '—'}</p>
      </div>

      {/* Progress stepper */}
      <Card>
        <CardContent className="p-6">
          <div className="mb-4 flex items-center justify-between">
            <h2 className="text-sm font-semibold">Progress Skripsi Anda</h2>
            <div className="flex items-center gap-2">
              <Badge variant={d.status === 'cancelled' ? 'danger' : d.status === 'graduated' ? 'success' : 'primary'}>
                {d.current_stage || d.status.replace(/_/g, ' ')}
              </Badge>
              <Badge variant="muted">{d.progress_percentage}%</Badge>
            </div>
          </div>
          {/* Mobile: horizontal snap scroll | Desktop: flex row */}
          <div className="-mx-1 overflow-x-auto pb-2 sm:mx-0 sm:overflow-visible sm:pb-0">
            <div className="flex min-w-[640px] items-center sm:min-w-0">
              {STAGE_ORDER.map((stage, i) => {
                const done = i <= stageIndex
                return (
                  <div key={stage} className="flex flex-1 items-center last:flex-none">
                    <div className="flex flex-col items-center gap-1.5">
                      <div
                        className={`flex h-8 w-8 shrink-0 items-center justify-center rounded-full text-xs font-bold ${
                          done ? 'bg-success text-white' : 'bg-muted text-muted-foreground'
                        }`}
                      >
                        {i + 1}
                      </div>
                      <span className="whitespace-nowrap text-[10px] text-muted-foreground">{STAGE_LABELS[stage]}</span>
                    </div>
                    {i < STAGE_ORDER.length - 1 && (
                      <div className={`mx-2 h-0.5 flex-1 rounded ${done ? 'bg-success' : 'bg-muted'}`} />
                    )}
                  </div>
                )
              })}
            </div>
          </div>
        </CardContent>
      </Card>

      {pendingActions.length > 0 && (
        <Card className="border-warning/30 bg-warning-50">
          <CardContent className="p-4">
            <p className="mb-2 flex items-center gap-2 text-sm font-semibold text-warning">
              <AlertCircle className="h-4 w-4" /> Yang perlu Anda lakukan
            </p>
            <ul className="space-y-1 text-sm">
              {pendingActions.map((a, i) => (
                <li key={i}>• {a}</li>
              ))}
            </ul>
          </CardContent>
        </Card>
      )}

      <div className="grid grid-cols-1 gap-4 sm:grid-cols-3">
        <StatCard title="Bimbingan" value={d.consultation_count ?? 0} icon={MessagesSquare} href="/supervision" iconClass="bg-primary-50 text-primary" />
        <StatCard title="Dokumen" value={totalDocs} icon={FolderOpen} href="/documents" iconClass="bg-secondary-50 text-secondary" />
        <StatCard
          title="Ujian Mendatang"
          value={(d.upcoming_seminar ? 1 : 0) + (d.upcoming_defense ? 1 : 0)}
          icon={CalendarDays}
          href="/schedules"
          iconClass="bg-success-50 text-success"
        />
      </div>

      <Card>
        <CardHeader>
          <CardTitle>Skripsi Saya</CardTitle>
          <CardDescription>Detail pengajuan skripsi Anda</CardDescription>
        </CardHeader>
        <CardContent className="space-y-3">
          <div>
            <p className="text-sm text-muted-foreground">Judul</p>
            <p className="font-medium">{d.title}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Pembimbing</p>
            <p className="font-medium">
              {d.supervisors?.length ? d.supervisors.map((s) => s.full_name).join(', ') : 'Belum ditentukan'}
            </p>
          </div>
          {d.last_consultation && (
            <div>
              <p className="text-sm text-muted-foreground">Bimbingan terakhir</p>
              <p className="font-medium">{formatDate(d.last_consultation)}</p>
            </div>
          )}
          {d.upcoming_seminar && (
            <div>
              <p className="text-sm text-muted-foreground">Seminar terjadwal</p>
              <p className="font-medium">
                {d.upcoming_seminar.scheduled_at ? formatDate(d.upcoming_seminar.scheduled_at) : '—'}
                {d.upcoming_seminar.room ? ` · ${d.upcoming_seminar.room}` : ''}
              </p>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  )
}
