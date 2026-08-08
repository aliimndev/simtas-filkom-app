import type { DocumentType } from '@/types/document'

// Document types match backend constants (Job 07/21).
export const DOCUMENT_TYPES: { value: DocumentType; label: string }[] = [
  { value: 'proposal', label: 'Proposal' },
  { value: 'draft_chapter', label: 'Draft Bab' },
  { value: 'seminar_doc', label: 'Dokumen Seminar' },
  { value: 'defense_doc', label: 'Dokumen Sidang' },
  { value: 'final_thesis', label: 'Skripsi Final' },
  { value: 'revision_sheet', label: 'Lembar Revisi' },
  { value: 'endorsement_letter', label: 'Surat Pengesahan' },
]
