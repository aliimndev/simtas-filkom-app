// Tipe dashboard ini diselaraskan persis dengan response shapes backend
// (backend/internal/usecase/dashboard_usecase.go, Job 12).

export interface DashboardSummary {
  academic_summary: {
    total_active: number
    total_graduated: number
    avg_completion_months: number
  }
  by_status: { status: string; label: string; count: number }[]
  graduation_trend: { month: string; count: number }[]
}

export interface ScheduleItem {
  id: string
  student_name: string
  thesis_title: string
  scheduled_at?: string | null
  room: string
}

export interface UpcomingSchedules {
  seminars: ScheduleItem[]
  defenses: ScheduleItem[]
}

export interface OperationalDashboard {
  pending_actions: {
    pending_title_reviews: number
    pending_document_reviews: number
    pending_seminars: number
    pending_defenses: number
  }
  upcoming_schedules: UpcomingSchedules
  activity_stats: {
    logins_today: number
    documents_uploaded_this_week: number
    consultations_this_week: number
  }
}

export interface StudentDashboard {
  thesis_id: string
  title: string
  status: string
  current_stage: string
  progress_percentage: number
  supervisors: { full_name: string; email?: string; nim?: string }[]
  documents: { type: string; status: string; version: number }[]
  consultation_count: number
  last_consultation?: string | null
  upcoming_seminar?: ScheduleInfo | null
  upcoming_defense?: ScheduleInfo | null
  pending_actions: string[]
}

export interface ScheduleInfo {
  id: string
  scheduled_at?: string | null
  room: string
  status: string
}

export interface SupervisedStudent {
  thesis_id: string
  student: { full_name: string; email?: string; nim?: string }
  title: string
  status: string
  current_stage: string
  pending_document_reviews: number
  last_consultation?: string | null
  consultation_count: number
  days_since_last_consultation: number
}

export interface SupervisorDashboard {
  total_students: number
  students: SupervisedStudent[]
  pending_document_reviews: number
  upcoming_schedules: UpcomingSchedules
}

export interface ExaminerAssignment {
  type: string
  id: string
  thesis_title: string
  student_name: string
  scheduled_at?: string | null
  room: string
  has_scored: boolean
}

export interface ExaminerDashboard {
  upcoming_assignments: ExaminerAssignment[]
  pending_scores: ExaminerAssignment[]
  scoring_history: ExaminerAssignment[]
}

/** Helper: gabungkan seminars + defenses menjadi satu daftar ber-tag. */
export function flattenSchedules(upcoming: UpcomingSchedules | undefined): (ScheduleItem & { type: 'seminar' | 'defense' })[] {
  if (!upcoming) return []
  return [
    ...(upcoming.seminars ?? []).map((s) => ({ ...s, type: 'seminar' as const })),
    ...(upcoming.defenses ?? []).map((d) => ({ ...d, type: 'defense' as const })),
  ].sort((a, b) => (a.scheduled_at ?? '').localeCompare(b.scheduled_at ?? ''))
}
