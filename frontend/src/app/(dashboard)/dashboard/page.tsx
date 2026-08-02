'use client'

import { useAuthStore } from '@/lib/stores/auth-store'
import AdminDashboardPage from './admin-page'
import StudentDashboardPage from './student-page'
import SupervisorDashboardPage from './supervisor-page'
import ExaminerDashboardPage from './examiner-page'
import { Spinner } from '@/components/ui/spinner'

export default function DashboardPage() {
  const { user } = useAuthStore()

  if (!user) return <Spinner label="Memuat dashboard…" />

  switch (user.role) {
    case 'mahasiswa':
      return <StudentDashboardPage />
    case 'dosen_pembimbing':
      return <SupervisorDashboardPage />
    case 'dosen_penguji':
      return <ExaminerDashboardPage />
    case 'admin_fakultas':
    case 'kaprodi':
      return <AdminDashboardPage />
    default:
      return <AdminDashboardPage />
  }
}
