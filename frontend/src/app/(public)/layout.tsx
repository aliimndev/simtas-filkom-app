'use client'

import Link from 'next/link'
import { useEffect, useState } from 'react'
import { ArrowRight, Menu, X } from 'lucide-react'
import { useAuthStore } from '@/lib/stores/auth-store'
import { BootLoader } from '@/components/features/landing/boot-loader'
import { cn } from '@/lib/utils/cn'

const NAV_LINKS = [
  { href: '/', label: 'Beranda' },
  { href: '/about', label: 'Tentang' },
  { href: '/guide', label: 'Panduan' },
  { href: '/faq', label: 'FAQ' },
  { href: '/contact', label: 'Kontak' },
]

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  const [open, setOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)
  const accessToken = useAuthStore((s) => s.accessToken)

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  return (
    <div className="simtas-dark flex min-h-screen flex-col">
      <BootLoader />

      {/* ── Floating pill navbar ─────────────────────────────────────── */}
      <header className="fixed inset-x-0 top-0 z-50 flex justify-center px-4 pt-4 md:pt-6">
        <div
          className={cn(
            'flex w-full max-w-3xl items-center justify-between rounded-full border border-st-stroke bg-(--st-surface)/80 px-2 py-2 backdrop-blur-md transition-shadow duration-300',
            scrolled && 'shadow-lg shadow-black/10',
          )}
        >
          {/* Logo */}
          <Link href="/" className="group flex items-center gap-2.5 pl-1">
            <span className="accent-ring relative flex h-9 w-9 items-center justify-center rounded-full">
              <span className="flex h-[calc(100%-3px)] w-[calc(100%-3px)] items-center justify-center rounded-full bg-st-bg">
                <span className="font-display text-[15px] italic text-st-text">
                  sf
                </span>
              </span>
            </span>
            <span className="hidden text-sm font-medium tracking-tight text-st-text sm:block">
              SIMTAS <span className="text-st-muted">FILKOM</span>
            </span>
          </Link>

          {/* Desktop nav */}
          <nav className="hidden items-center gap-0.5 md:flex">
            {NAV_LINKS.map((l) => (
              <Link
                key={l.href}
                href={l.href}
                className="rounded-full px-3.5 py-1.5 text-sm text-st-muted transition hover:bg-st-surface-hi hover:text-st-text"
              >
                {l.label}
              </Link>
            ))}
          </nav>

          <div className="flex items-center gap-1.5">
            {accessToken ? (
              <Link
                href="/dashboard"
                className="accent-ring hidden items-center gap-1.5 rounded-full border border-st-stroke bg-st-surface-hi px-4 py-1.5 text-sm text-st-text transition hover:text-st-text md:inline-flex"
              >
                Dashboard <ArrowRight className="h-3.5 w-3.5" />
              </Link>
            ) : (
              <Link
                href="/login"
                className="accent-ring hidden rounded-full bg-st-text px-4 py-1.5 text-sm font-medium text-st-bg transition hover:opacity-90 md:inline-block"
              >
                Masuk
              </Link>
            )}
            <button
              type="button"
              aria-label={open ? 'Tutup menu' : 'Buka menu'}
              onClick={() => setOpen((v) => !v)}
              className="inline-flex h-9 w-9 items-center justify-center rounded-full text-st-muted hover:bg-st-surface-hi hover:text-st-text md:hidden"
            >
              {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </button>
          </div>
        </div>
      </header>

      {open && (
        <nav className="fixed inset-x-0 top-20 z-40 mx-auto w-[calc(100%-2rem)] max-w-3xl rounded-2xl border border-st-stroke bg-st-surface p-2 md:hidden">
          {NAV_LINKS.map((l) => (
            <Link
              key={l.href}
              href={l.href}
              onClick={() => setOpen(false)}
              className="block rounded-xl px-4 py-3 text-sm text-st-muted hover:bg-st-surface-hi hover:text-st-text"
            >
              {l.label}
            </Link>
          ))}
          <Link
            href={accessToken ? '/dashboard' : '/login'}
            onClick={() => setOpen(false)}
            className="mt-1 flex items-center justify-between rounded-xl bg-st-text px-4 py-3 text-sm font-medium text-st-bg"
          >
            {accessToken ? 'Buka Dashboard' : 'Masuk ke Sistem'}
            <ArrowRight className="h-4 w-4" />
          </Link>
        </nav>
      )}

      <main className="flex-1">{children}</main>

      <footer className="border-t border-st-stroke">
        <div className="landing-container grid gap-12 py-16 md:grid-cols-4">
          <div className="md:col-span-2">
            <div className="flex items-center gap-2.5">
              <span className="accent-ring flex h-8 w-8 items-center justify-center rounded-full">
                <span className="flex h-[calc(100%-3px)] w-[calc(100%-3px)] items-center justify-center rounded-full bg-st-bg">
                  <span className="font-display text-[13px] italic text-st-text">
                    sf
                  </span>
                </span>
              </span>
              <span className="text-sm font-medium tracking-tight text-st-text">
                SIMTAS <span className="text-st-muted">FILKOM</span>
              </span>
            </div>
            <p className="mt-5 max-w-sm text-sm leading-relaxed text-st-muted">
              Sistem Manajemen Tugas Akhir Skripsi Fakultas Ilmu Komputer
              Universitas Djuanda — mengelola seluruh proses Tugas Akhir dari
              pengajuan judul, bimbingan, seminar, hingga sidang dan arsip.
            </p>
          </div>

          <div>
            <h4 className="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-st-muted">
              Navigasi
            </h4>
            <ul className="mt-5 space-y-3 text-sm">
              {NAV_LINKS.map((l) => (
                <li key={l.href}>
                  <Link href={l.href} className="text-st-muted transition hover:text-st-text">
                    {l.label}
                  </Link>
                </li>
              ))}
            </ul>
          </div>

          <div>
            <h4 className="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-st-muted">
              Akses Sistem
            </h4>
            <ul className="mt-5 space-y-3 text-sm">
              <li>
                <Link href="/login" className="text-st-muted transition hover:text-st-text">
                  Masuk
                </Link>
              </li>
              <li>
                <Link href="/forgot-password" className="text-st-muted transition hover:text-st-text">
                  Lupa Password
                </Link>
              </li>
            </ul>
          </div>
        </div>

        <div className="border-t border-st-stroke">
          <div className="landing-container flex flex-col items-center justify-between gap-2 py-6 text-xs text-st-muted sm:flex-row">
            <p>© {new Date().getFullYear()} SIMTAS FILKOM — Universitas Djuanda.</p>
            <p>Fakultas Ilmu Komputer · Universitas Djuanda</p>
          </div>
        </div>
      </footer>
    </div>
  )
}
