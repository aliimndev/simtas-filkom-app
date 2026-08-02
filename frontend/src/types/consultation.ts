export interface Consultation {
  id: string
  thesis_id: string
  student_id?: string
  student_name?: string
  supervisor_id?: string
  supervisor_name?: string
  date: string
  topic: string
  notes?: string
  follow_up?: string
  status: 'pending' | 'approved' | 'rejected'
  created_at?: string
}

export interface CreateConsultationRequest {
  date: string
  topic: string
  notes?: string
  follow_up?: string
}

export interface UpdateConsultationRequest {
  date?: string
  topic?: string
  notes?: string
  follow_up?: string
}

export interface ConsultationSummary {
  total: number
  approved: number
  pending: number
  rejected: number
  last_consultation_at?: string
}
