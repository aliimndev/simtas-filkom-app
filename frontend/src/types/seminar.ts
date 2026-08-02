export type SeminarStatus = 'submitted' | 'scheduled' | 'in_revision' | 'passed' | 'failed'

export interface SeminarExaminer {
  id: string
  full_name: string
  nim_nidn?: string
}

export interface Seminar {
  id: string
  thesis_id: string
  thesis_title?: string
  student?: { id: string; full_name: string; nim_nidn?: string }
  stage: 'proposal' | 'progress' | 'final'
  status: SeminarStatus
  scheduled_at?: string
  room?: string
  examiners?: SeminarExaminer[]
  scores?: Record<string, number>
  average_score?: number
  revision_notes?: string
  submitted_at?: string
  created_at?: string
}

export interface SubmitSeminarRequest {
  stage: 'proposal' | 'progress' | 'final'
  scheduled_at?: string
}

export interface ScheduleSeminarRequest {
  scheduled_at: string
  room: string
  examiner_ids: string[]
}

export interface SeminarScore {
  examiner_id: string
  score: number
  notes?: string
}

export interface SubmitScoresRequest {
  scores: SeminarScore[]
}
