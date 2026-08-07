'use client'

import * as React from 'react'
import { FileText, UploadCloud, X } from 'lucide-react'
import { cn } from '@/lib/utils/cn'
import { formatFileSize, validatePdfFile } from '@/lib/utils/file'

export interface FileDropzoneProps {
  id?: string
  value?: File | null
  onChange: (file: File | null) => void
  invalid?: boolean
  accept?: string
  maxSizeBytes?: number
  className?: string
}

/**
 * Single-file dropzone for the draft proposal upload.
 * Validates type/size locally; non-PDF and oversize files are rejected
 * inline and the selected file is cleared.
 */
export function FileDropzone({
  id,
  value,
  onChange,
  invalid,
  accept = 'application/pdf,.pdf',
  maxSizeBytes,
  className,
}: FileDropzoneProps) {
  const inputRef = React.useRef<HTMLInputElement>(null)
  const [dragOver, setDragOver] = React.useState(false)
  const [error, setError] = React.useState<string | null>(null)

  const handleFile = (file: File | undefined) => {
    if (!file) return
    let message = validatePdfFile(file)
    if (!message && maxSizeBytes && file.size > maxSizeBytes) {
      message = `Ukuran file maksimal ${formatFileSize(maxSizeBytes)}.`
    }
    if (message) {
      setError(message)
      onChange(null)
      return
    }
    setError(null)
    onChange(file)
  }

  return (
    <div className="space-y-1">
      <div
        role="button"
        tabIndex={0}
        aria-disabled={false}
        onClick={() => inputRef.current?.click()}
        onKeyDown={(e) => {
          if (e.key === 'Enter' || e.key === ' ') {
            e.preventDefault()
            inputRef.current?.click()
          }
        }}
        onDragOver={(e) => {
          e.preventDefault()
          setDragOver(true)
        }}
        onDragLeave={() => setDragOver(false)}
        onDrop={(e) => {
          e.preventDefault()
          setDragOver(false)
          handleFile(e.dataTransfer.files[0])
        }}
        className={cn(
          'flex min-h-[120px] w-full cursor-pointer flex-col items-center justify-center gap-1 rounded-lg border border-dashed bg-background px-3 py-6 text-center transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring',
          dragOver ? 'border-primary bg-primary-50' : 'border-border',
          invalid || error ? 'border-danger focus-visible:ring-danger' : 'border-border',
          className,
        )}
      >
        <input
          ref={inputRef}
          id={id}
          type="file"
          accept={accept}
          className="sr-only"
          onChange={(e) => {
            handleFile(e.target.files?.[0])
            e.target.value = ''
          }}
        />
        {value ? (
          <div className="flex w-full items-center justify-center gap-2 text-sm">
            <FileText className="h-5 w-5 shrink-0 text-primary" />
            <span className="min-w-0 truncate font-medium">{value.name}</span>
            <span className="shrink-0 text-xs text-muted-foreground">
              {formatFileSize(value.size)}
            </span>
          </div>
        ) : (
          <>
            <UploadCloud className="h-8 w-8 text-muted-foreground" aria-hidden />
            <p className="text-sm font-medium text-muted-foreground">
              Pilih file PDF atau drag &amp; drop di sini
            </p>
            <p className="text-xs text-muted-foreground">PDF, maksimal 10 MB</p>
          </>
        )}
      </div>
      {value && (
        <button
          type="button"
          onClick={() => {
            onChange(null)
            setError(null)
          }}
          className="inline-flex items-center gap-1 text-xs text-muted-foreground hover:text-danger"
        >
          <X className="h-3.5 w-3.5" /> Hapus file
        </button>
      )}
      {(error || invalid) && (
        <p className="mt-1 text-xs text-danger">{error ?? 'Draft proposal wajib diunggah'}</p>
      )}
    </div>
  )
}
