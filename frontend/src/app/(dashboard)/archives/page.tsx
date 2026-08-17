'use client'

import { useQuery } from '@tanstack/react-query'
import { Search, Download, FileArchive, Loader2 } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { ListSkeleton } from '@/components/ui/skeleton'
import { archiveApi } from '@/lib/api/archive-api'
import { formatDate } from '@/lib/utils/date'
import { getErrorMessage } from '@/lib/utils/error'

export default function ArchivesPage() {
  const [q, setQ] = useState('')
  const [page, setPage] = useState(1)
  const [downloadingId, setDownloadingId] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const archives = useQuery({
    queryKey: ['archives', q, page],
    queryFn: () => archiveApi.search({ q: q || undefined, page, per_page: 12 }),
  })

  const list = archives.data?.data ?? []
  const total = archives.data?.meta.total ?? 0

  async function handleDownload(id: string) {
    setDownloadingId(id)
    setError(null)
    try {
      await archiveApi.download(id)
    } catch (e) {
      setError(getErrorMessage(e))
    } finally {
      setDownloadingId(null)
    }
  }

  return (
    <div className="space-y-6">
      <div>
        <p className="landing-eyebrow">Arsip Skripsi</p>
        <h1 className="mt-2 text-balance landing-heading text-2xl">
          Kumpulan skripsi yang telah <span className="accent-text italic">diarsipkan</span>
        </h1>
      </div>

      <div className="relative max-w-md">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Cari judul / mahasiswa…"
          value={q}
          onChange={(e) => { setQ(e.target.value); setPage(1) }}
          className="pl-9"
        />
      </div>

      {error && <p className="text-sm text-danger">{error}</p>}

      {archives.isLoading ? (
        <ListSkeleton count={6} label="Memuat arsip…" />
      ) : (
        <>
          <div className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3">
            {list.map((a) => (
              <Card key={a.id} className="group">
                <CardContent className="p-5">
                  <div className="mb-3 flex items-center justify-between">
                    <div className="accent-ring flex h-10 w-10 items-center justify-center rounded-xl bg-secondary-50 text-secondary">
                      <FileArchive className="h-5 w-5" />
                    </div>
                    <Button
                      size="sm"
                      variant="ghost"
                      disabled={downloadingId === a.id}
                      onClick={() => handleDownload(a.id)}
                    >
                      {downloadingId === a.id ? <Loader2 className="h-4 w-4 animate-spin" /> : <Download className="h-4 w-4" />}
                    </Button>
                  </div>
                  <p className="line-clamp-2 font-medium leading-snug">{a.title ?? 'Skripsi'}</p>
                  <p className="mt-1 text-sm text-muted-foreground">{a.student?.full_name ?? '—'}</p>
                  {a.graduation_year ? (
                    <p className="mt-2 text-xs text-muted-foreground">Tahun {a.graduation_year} · Diarsipkan {formatDate(a.archived_at)}</p>
                  ) : (
                    <p className="mt-2 text-xs text-muted-foreground">Diarsipkan {formatDate(a.archived_at)}</p>
                  )}
                </CardContent>
              </Card>
            ))}
          </div>

          {list.length === 0 && (
            <div className="py-12 sm:py-12 text-center">
              <FileArchive className="mx-auto h-10 w-10 text-st-muted" />
              <p className="mt-3 landing-heading text-lg">Tidak ada <span className="accent-text italic">arsip</span> ditemukan</p>
              <p className="mt-1 text-sm text-st-muted">Coba ubah kata kunci pencarian Anda.</p>
            </div>
          )}

          {total > 12 && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">Total {total} arsip · Halaman {page}</p>
              <div className="flex gap-2">
                <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Sebelumnya</Button>
                <Button size="sm" variant="outline" disabled={page >= Math.ceil(total / 12)} onClick={() => setPage((p) => p + 1)}>Berikutnya</Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
