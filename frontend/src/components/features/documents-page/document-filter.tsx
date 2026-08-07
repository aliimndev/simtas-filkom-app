import { Button } from '@/components/ui/button'
import { DOCUMENT_TYPES } from './document-types'

/** Filter dokumen berdasarkan jenis (Semua / per jenis). */
export function DocumentFilter({
  selectedType,
  onSelect,
}: {
  selectedType: string
  onSelect: (type: string) => void
}) {
  return (
    <div className="flex flex-wrap gap-2">
      <Button
        size="sm"
        variant={selectedType === '' ? 'primary' : 'outline'}
        onClick={() => onSelect('')}
      >
        Semua
      </Button>
      {DOCUMENT_TYPES.map((t) => (
        <Button
          key={t.value}
          size="sm"
          variant={selectedType === t.value ? 'primary' : 'outline'}
          onClick={() => onSelect(t.value)}
        >
          {t.label}
        </Button>
      ))}
    </div>
  )
}
