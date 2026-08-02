export type DefenseStatus = 'submitted' | 'scheduled' | 'in_revision' | 'passed' | 'failed'

export interface DefenseExaminer {
  id: string
  full_name: string
  nim_nidn?: string
}

export interface Defense {
  id: string
  thesis_id: string
  thesis_title?: string
  student?: { id: string; full_name: string; nim_nidn?: string }
  status: DefenseStatus
  scheduled_at?: string
  room?: string
  examiners?: DefenseExaminer[]
  scores?: Record<string, number>
  average_score?: number
  revision_notes?: string
  submitted_at?: string
  created_at?: string
}

export interface ScheduleDefenseRequest {
  scheduled_at: string
  room: string
  examiner_ids: string[]
}

export interface DefenseScore {
  examiner_id: string
  score: number
  notes?: string
}

export interface SubmitDefenseScoresRequest {
  scores: DefenseScore[]
}

export interface GraduationRequest {
  decision: 'graduated' | 'rejected'
  notes?: string
}
