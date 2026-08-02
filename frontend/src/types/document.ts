// Keep in sync with backend/internal/domain/entity/document_types.go (Job 07/21).
export type DocumentType =
  | 'proposal'
  | 'draft_chapter'
  | 'seminar_doc'
  | 'defense_doc'
  | 'final_thesis'
  | 'revision_sheet'
  | 'endorsement_letter'

export type DocumentStatus = 'pending_review' | 'approved' | 'revision_required'

export interface ThesisDocument {
  id: string
  thesis_id?: string
  document_type: DocumentType
  chapter_number?: number | null
  version: number
  file_name: string
  file_size?: number | null
  status: DocumentStatus
  reviewer?: { id: string; full_name: string } | null
  reviewer_notes?: string | null
  reviewed_at?: string | null
  created_at?: string
  updated_at?: string
}

export const DOCUMENT_TYPE_LABELS: Record<DocumentType, string> = {
  proposal: 'Proposal',
  draft_chapter: 'Draft Bab',
  seminar_doc: 'Dokumen Seminar',
  defense_doc: 'Dokumen Sidang',
  final_thesis: 'Skripsi Final',
  revision_sheet: 'Lembar Revisi',
  endorsement_letter: 'Surat Pengesahan',
}

export const DOCUMENT_STATUS_LABELS: Record<DocumentStatus, string> = {
  pending_review: 'Menunggu Review',
  approved: 'Disetujui',
  revision_required: 'Perlu Revisi',
}

export function documentTypeLabel(type?: string): string {
  if (!type) return '—'
  return DOCUMENT_TYPE_LABELS[type as DocumentType] ?? type
}
