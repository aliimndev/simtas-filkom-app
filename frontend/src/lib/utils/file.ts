export const MAX_PDF_SIZE_BYTES = 10 * 1024 * 1024 // 10 MB, sync with backend MaxDocumentSizeBytes

export const PDF_ERRORS = {
  notPdf: 'File harus berformat PDF.',
  tooLarge: 'Ukuran file maksimal 10 MB.',
} as const

export function validatePdfFile(file: File): string | null {
  if (!/\.pdf$/i.test(file.name)) return PDF_ERRORS.notPdf
  if (file.size > MAX_PDF_SIZE_BYTES) return PDF_ERRORS.tooLarge
  return null
}

export function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}
