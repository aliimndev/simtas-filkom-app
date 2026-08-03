export type ThesisStatus =
  | 'submitted'
  | 'approved'
  | 'rejected'
  | 'in_progress'
  | 'seminar_ready'
  | 'seminar_done'
  | 'defense_ready'
  | 'defense_done'
  | 'graduated'
  | 'cancelled'

export interface ThesisSupervisor {
  id: string
  full_name: string
  nim_nidn?: string
  email?: string
}

export interface ThesisStudent {
  id: string
  full_name: string
  nim_nidn?: string
  email?: string
}

export interface Thesis {
  id: string
  title: string
  abstract?: string
  field_of_study?: string
  status: ThesisStatus
  student?: ThesisStudent
  student_id?: string
  supervisors?: ThesisSupervisor[]
  academic_year_id?: string
  academic_year?: { id: string; name?: string }
  submitted_at: string
  approved_at?: string
  graduated_at?: string
  kaprodi_notes?: string
}

export interface ThesisFilter {
  status?: ThesisStatus
  q?: string
  page?: number
  per_page?: number
}

export interface CreateThesisRequest {
  title: string
  abstract: string
  field_of_study: string
  thesis_type: 'skripsi' | 'tugas_akhir'
  academic_year_id: string
}

export interface ReviewThesisRequest {
  decision: 'approved' | 'rejected'
  notes?: string
}

export interface AssignSupervisorRequest {
  supervisor_ids: string[]
}
