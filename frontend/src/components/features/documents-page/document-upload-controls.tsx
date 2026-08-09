import { Upload, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Label } from '@/components/ui/label'
import { Select } from '@/components/ui/select'
import { DOCUMENT_TYPES } from './document-types'
import type { DocumentType } from '@/types/document'

/** Kontrol unggah dokumen (khusus mahasiswa): jenis, bab, dan tombol unggah. */
export function DocumentUploadControls({
  uploadType,
  onUploadTypeChange,
  chapterNumber,
  onChapterNumberChange,
  uploading,
  onUploadFile,
  onCancelUpload,
}: {
  uploadType: DocumentType
  onUploadTypeChange: (type: DocumentType) => void
  chapterNumber: string
  onChapterNumberChange: (chapter: string) => void
  uploading: boolean
  onUploadFile: (file: File) => void
  onCancelUpload: () => void
}) {
  return (
    <div className="flex flex-wrap items-end gap-2">
      <div>
        <Label htmlFor="type">Jenis Dokumen</Label>
        <Select id="type" value={uploadType} onChange={(e) => onUploadTypeChange(e.target.value as DocumentType)} className="w-44">
          {DOCUMENT_TYPES.map((t) => (
            <option key={t.value} value={t.value}>{t.label}</option>
          ))}
        </Select>
      </div>
      {uploadType === 'draft_chapter' && (
        <div>
          <Label htmlFor="chapter">Bab</Label>
          <Select id="chapter" value={chapterNumber} onChange={(e) => onChapterNumberChange(e.target.value)} className="w-20">
            {['1', '2', '3', '4', '5'].map((n) => (
              <option key={n} value={n}>{n}</option>
            ))}
          </Select>
        </div>
      )}
      {uploading ? (
        <Button size="sm" variant="ghost" onClick={onCancelUpload}>
          <X className="h-4 w-4" /> Batalkan
        </Button>
      ) : (
        <label className="inline-flex h-10 cursor-pointer items-center gap-2 rounded-lg bg-primary px-4 text-sm font-medium text-primary-foreground shadow-sm transition-colors hover:bg-primary-700">
          <Upload className="h-4 w-4" />
          Unggah
          <input
            type="file"
            accept=".pdf"
            className="hidden"
            onChange={(e) => {
              const f = e.target.files?.[0]
              if (f) onUploadFile(f)
              e.target.value = ''
            }}
          />
        </label>
      )}
    </div>
  )
}
