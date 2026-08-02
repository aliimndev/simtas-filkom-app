import apiClient from './client'
import type { DashboardSummary, StudentDashboard, SupervisorDashboard, ExaminerDashboard, OperationalDashboard } from '@/types/dashboard'

export const dashboardApi = {
  async summary(params?: { academic_year_id?: string }): Promise<DashboardSummary> {
    const res = await apiClient.get('/dashboard/summary', { params })
    return res.data.data
  },

  async operational(params?: { days?: number }): Promise<OperationalDashboard> {
    const res = await apiClient.get('/dashboard/operational', { params })
    return res.data.data
  },

  async student(): Promise<StudentDashboard> {
    const res = await apiClient.get('/dashboard/student')
    return res.data.data
  },

  async supervisor(): Promise<SupervisorDashboard> {
    const res = await apiClient.get('/dashboard/supervisor')
    return res.data.data
  },

  async examiner(): Promise<ExaminerDashboard> {
    const res = await apiClient.get('/dashboard/examiner')
    return res.data.data
  },
}
