import { BookOpen } from 'lucide-react'
import Link from 'next/link'
import { Button } from '@/components/ui/button'

/** Empty state: mahasiswa belum memiliki skripsi aktif. */
export function NoThesisState() {
  return (
    <div className="mx-auto max-w-lg space-y-6 py-10 text-center">
      <div className="mx-auto flex h-16 w-16 items-center justify-center rounded-2xl bg-primary-50 text-primary">
        <BookOpen className="h-8 w-8" />
      </div>
      <div>
        <h1 className="text-xl font-bold">Anda belum memiliki skripsi</h1>
        <p className="mt-2 text-sm text-muted-foreground">
          Ajukan judul skripsi Anda untuk memulai perjalanan Tugas Akhir.
        </p>
      </div>
      <Button asChild size="lg">
        <Link href="/thesis/new">Ajukan Judul Skripsi</Link>
      </Button>
    </div>
  )
}
