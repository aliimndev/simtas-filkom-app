import { Button } from '@/components/ui/button'

/** Kontrol pagination sederhana (sebelumnya / berikutnya). */
export function Pagination({
  page,
  total,
  perPage,
  onPageChange,
}: {
  page: number
  total: number
  perPage: number
  onPageChange: (page: number) => void
}) {
  if (total <= perPage) return null

  return (
    <div className="flex items-center justify-between">
      <p className="text-sm text-muted-foreground">Total {total} pengguna · Halaman {page}</p>
      <div className="flex gap-2">
        <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => onPageChange(page - 1)}>Sebelumnya</Button>
        <Button size="sm" variant="outline" disabled={page >= Math.ceil(total / perPage)} onClick={() => onPageChange(page + 1)}>Berikutnya</Button>
      </div>
    </div>
  )
}
