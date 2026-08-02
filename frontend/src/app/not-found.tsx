import Link from 'next/link'
import { Compass, ArrowRight, GraduationCap } from 'lucide-react'

export default function NotFound() {
  return (
    <div className="landing-grid-bg flex min-h-screen flex-col">
      <div className="landing-container flex flex-1 flex-col items-center justify-center py-24 text-center">
        <span className="flex h-14 w-14 items-center justify-center rounded-2xl bg-primary text-primary-foreground">
          <Compass className="h-7 w-7" />
        </span>
        <p className="mt-8 text-sm font-semibold uppercase tracking-[0.2em] text-primary">Error 404</p>
        <h1 className="landing-display mt-3 text-5xl md:text-6xl">Halaman tidak ditemukan</h1>
        <p className="mt-5 max-w-md text-muted-foreground">
          Halaman yang Anda cari mungkin telah dipindahkan, dihapus, atau alamatnya salah.
        </p>
        <div className="mt-9 flex flex-col items-center gap-3 sm:flex-row">
          <Link
            href="/"
            className="inline-flex items-center gap-2 rounded-lg bg-primary px-6 py-3 text-sm font-medium text-primary-foreground transition hover:bg-primary-700"
          >
            Kembali ke Beranda <ArrowRight className="h-4 w-4" />
          </Link>
          <Link
            href="/login"
            className="inline-flex items-center gap-2 rounded-lg border border-border px-6 py-3 text-sm font-medium text-foreground transition hover:bg-muted"
          >
            <GraduationCap className="h-4 w-4" /> Masuk
          </Link>
        </div>
      </div>
    </div>
  )
}
