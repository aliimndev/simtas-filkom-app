// Keep in sync with backend/internal/usecase/archive_usecase.go ArchiveDetail (Job 10/21).
export interface Archive {
  id: string
  thesis_id: string
  title: string
  abstract_id?: string
  abstract_en?: string | null
  keywords?: string[]
  graduation_year?: number
  field_of_study?: string
  study_program?: string
  student?: { id: string; full_name: string; nim?: string }
  supervisors?: { id: string; full_name: string }[]
  file_name?: string
  archived_at?: string
  created_at?: string
}

export interface ArchiveStats {
  total_archives: number
  by_year?: { year: number; count: number }[]
  by_field?: { field: string; count: number }[]
  by_study_program?: { program: string; count: number }[]
}
