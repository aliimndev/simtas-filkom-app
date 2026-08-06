export type TitleChangeStatus = 'PENDING' | 'APPROVED' | 'REJECTED' | 'CANCELLED'

export interface TitleChangeUserBrief {
  id: string
  full_name: string
  nim_nidn?: string
}

export interface TitleChangeRequest {
  id: string
  thesis_id: string
  previous_title: string
  requested_title: string
  reason?: string
  status: TitleChangeStatus
  requested_by?: TitleChangeUserBrief
  reviewed_by?: TitleChangeUserBrief
  review_notes?: string
  cancelled_by?: TitleChangeUserBrief
  cancelled_at?: string
  created_at: string
  updated_at: string
}

export interface CreateTitleChangeRequest {
  requested_title: string
  reason?: string
}

export interface ReviewTitleChangeRequest {
  review_notes?: string
}
