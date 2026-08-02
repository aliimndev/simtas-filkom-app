'use client'

import Link from 'next/link'
import { GraduationCap, Menu, X, ArrowRight } from 'lucide-react'
import { useState } from 'react'
import { useAuthStore } from '@/lib/stores/auth-store'

const NAV_LINKS = [
  { href: '/', label: 'Beranda' },
  { href: '/about', label: 'Tentang' },
  { href: '/guide', label: 'Panduan' },
  { href: '/faq', label: 'FAQ' },
  { href: '/contact', label: 'Kontak' },
]

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState(false)
  const accessToken = useAuthStore((s) => s.accessToken)

  return (
    <div className="flex min-h-screen flex-col bg-background">
      <header className="sticky top-0 z-40 border-b border-border bg-background/80 backdrop-blur">
        <div className="landing-container flex h-16 items-center justify-between">
          <Link href="/" className="flex items-center gap-2.5">
            <span className="flex h-9 w-9 items-center justify-center rounded-lg bg-primary text-primary-foreground">
              <GraduationCap className="h-5 w-5" />
            </span>
            <span className="text-sm font-bold tracking-tight text-foreground">
              SIMTAS <span className="text-primary">FILKOM</span>
            </span>
          </Link>

          <nav className="hidden items-center gap-1 md:flex">
            {NAV_LINKS.map((l) => (
              <Link
                key={l.href}
                href={l.href}
                className="rounded-md px-3 py-2 text-sm font-medium text-muted-foreground transition hover:bg-muted hover:text-foreground"
              >
                {l.label}
              </Link>
            ))}
          </nav>

          <div className="hidden items-center gap-2 md:flex">
            {accessToken ? (
              <Link
                href="/dashboard"
                className="inline-flex items-center gap-1.5 rounded-lg bg-primary px-4 py-2 text-sm font-medium text-primary-foreground transition hover:bg-primary-700"
              >
                Dashboard <ArrowRight className="h-4 w-4" />
              </Link>
            ) : (
              <Link
                href="/login"
                className="inline-flex items-center gap-1.5 rounded-lg border border-border px-4 py-2 text-sm font-medium text-foreground transition hover:bg-muted"
              >
                Masuk
              </Link>
            )}
          </div>

          <button
            type="button"
            aria-label={open ? 'Tutup menu' : 'Buka menu'}
            onClick={() => setOpen((v) => !v)}
            className="inline-flex h-10 w-10 items-center justify-center rounded-lg text-muted-foreground hover:bg-muted md:hidden"
          >
            {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
          </button>
        </div>

        {open && (
          <nav className="border-t border-border bg-background md:hidden">
            <div className="landing-container flex flex-col gap-1 py-3">
              {NAV_LINKS.map((l) => (
                <Link
                  key={l.href}
                  href={l.href}
                  onClick={() => setOpen(false)}
                  className="rounded-md px-3 py-2.5 text-sm font-medium text-muted-foreground hover:bg-muted hover:text-foreground"
                >
                  {l.label}
                </Link>
              ))}
              <Link
                href={accessToken ? '/dashboard' : '/login'}
                onClick={() => setOpen(false)}
                className="mt-2 inline-flex items-center justify-center gap-1.5 rounded-lg bg-primary px-4 py-2.5 text-sm font-medium text-primary-foreground"
              >
                {accessToken ? 'Dashboard' : 'Masuk'} <ArrowRight className="h-4 w-4" />
              </Link>
            </div>
          </nav>
        )}
      </header>

      <main className="flex-1">{children}</main>

      <footer className="border-t border-border bg-muted/40">
        <div className="landing-container grid gap-10 py-14 md:grid-cols-4">
          <div className="md:col-span-2">
            <div className="flex items-center gap-2.5">
              <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
                <GraduationCap className="h-4 w-4" />
              </span>
              <span className="text-sm font-bold tracking-tight text-foreground">
                SIMTAS <span className="text-primary">FILKOM</span>
              </span>
            </div>
            <p className="mt-4 max-w-sm text-sm leading-relaxed text-muted-foreground">
              Sistem Informasi Manajemen Tugas Akhir &amp; Skripsi Fakultas Ilmu Komputer Universitas Djuanda —
              mengelola seluruh proses Tugas Akhir dari pengajuan judul, bimbingan, seminar, hingga sidang dan arsip.
            </p>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-foreground">Navigasi</h4>
            <ul className="mt-4 space-y-2.5 text-sm">
              {NAV_LINKS.map((l) => (
                <li key={l.href}>
                  <Link href={l.href} className="text-muted-foreground transition hover:text-primary">
                    {l.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h4 className="text-sm font-semibold text-foreground">Akses Sistem</h4>
            <ul className="mt-4 space-y-2.5 text-sm">
              <li>
                <Link href="/login" className="text-muted-foreground transition hover:text-primary">
                  Masuk
                </Link>
              </li>
              <li>
                <Link href="/forgot-password" className="text-muted-foreground transition hover:text-primary">
                  Lupa Password
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-border">
          <div className="landing-container flex flex-col items-center justify-between gap-2 py-6 text-xs text-muted-foreground sm:flex-row">
            <p>© {new Date().getFullYear()} SIMTAS FILKOM — Universitas Djuanda. Hak cipta dilindungi.</p>
            <p>Fakultas Ilmu Komputer · Universitas Djuanda</p>
          </div>
        </div>
      </footer>
    </div>
  )
}
