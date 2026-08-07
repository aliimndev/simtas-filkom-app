import { MAX_PDF_SIZE_BYTES, formatFileSize, validatePdfFile } from './file'

describe('validatePdfFile', () => {
  it('accepts a pdf under 10 MB', () => {
    const file = new File([new Uint8Array(1024)], 'draft.pdf', { type: 'application/pdf' })
    expect(validatePdfFile(file)).toBeNull()
  })

  it('rejects non-pdf extensions', () => {
    const file = new File(['x'], 'draft.docx', { type: 'application/vnd.openxmlformats-officedocument.wordprocessingml.document' })
    expect(validatePdfFile(file)).toBe('File harus berformat PDF.')
  })

  it('rejects a file over 10 MB', () => {
    const file = new File([new Uint8Array(MAX_PDF_SIZE_BYTES + 1)], 'draft.pdf', { type: 'application/pdf' })
    expect(validatePdfFile(file)).toBe('Ukuran file maksimal 10 MB.')
  })
})

describe('formatFileSize', () => {
  it('formats bytes, KB, and MB', () => {
    expect(formatFileSize(500)).toBe('500 B')
    expect(formatFileSize(2048)).toBe('2.0 KB')
    expect(formatFileSize(5 * 1024 * 1024)).toBe('5.0 MB')
  })
})
