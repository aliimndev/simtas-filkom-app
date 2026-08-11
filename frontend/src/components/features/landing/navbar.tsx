'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { useEffect, useState } from 'react'
import { useTheme } from '@/providers/theme-provider'
import { ArrowRight, Menu, X } from 'lucide-react'
import { useAuthStore } from '@/lib/stores/auth-store'
import { AnimatedThemeToggler } from '@/components/ui/animated-theme-toggler'
import { cn } from '@/lib/utils/cn'

type NavLink = { href: string; label: string }

export function NavbarSection({ navLinks }: { navLinks: NavLink[] }) {
  const [open, setOpen] = useState(false)
  const [scrolled, setScrolled] = useState(false)
  const pathname = usePathname()
  const accessToken = useAuthStore((s) => s.accessToken)
  const { theme, setTheme } = useTheme()

  useEffect(() => {
    const onScroll = () => setScrolled(window.scrollY > 40)
    onScroll()
    window.addEventListener('scroll', onScroll, { passive: true })
    return () => window.removeEventListener('scroll', onScroll)
  }, [])

  // Close the mobile menu on browser back/forward (popstate) so it can
  // never linger over a page the user navigated away from.
  useEffect(() => {
    const onPop = () => setOpen(false)
    window.addEventListener('popstate', onPop)
    return () => window.removeEventListener('popstate', onPop)
  }, [])

  // Close the mobile menu with the Escape key while it is open.
  useEffect(() => {
    if (!open) return
    const onKey = (e: KeyboardEvent) => {
      if (e.key === 'Escape') setOpen(false)
    }
    window.addEventListener('keydown', onKey)
    return () => window.removeEventListener('keydown', onKey)
  }, [open])

  return (
    <header className="fixed inset-x-0 top-0 z-50 flex justify-center px-4 pt-4 md:pt-6">
      {/* Relative wrapper so the mobile dropdown always sits directly below
          the pill — regardless of safe-area insets (notch) that push the
          header down — and can never be covered by it. */}
      <div
        className={cn(
          'relative w-full transition-all duration-300 ease-out',
          scrolled ? 'max-w-3xl' : 'max-w-5xl',
        )}
      >
        <div
          className={cn(
            'flex w-full items-center justify-between rounded-full border py-2 transition-all duration-300 ease-out',
            scrolled
              ? 'border-st-stroke bg-(--st-surface)/70 px-2 shadow-sm shadow-black/5 backdrop-blur-md'
              : 'border-transparent bg-transparent px-3',
          )}
        >
          {/* Logo */}
          <Link href="/" className="group flex min-w-0 items-center gap-2.5 pl-1">
            <span className="accent-ring relative flex h-9 w-9 shrink-0 items-center justify-center rounded-full">
              <span className="flex h-[calc(100%-3px)] w-[calc(100%-3px)] items-center justify-center rounded-full bg-st-bg">
                <span className="font-display text-[15px] italic text-st-text">
                  sf
                </span>
              </span>
            </span>
            <span className="hidden truncate text-sm font-medium tracking-tight text-st-text sm:block">
              SIMTAS <span className="text-st-muted">FILKOM</span>
            </span>
          </Link>

          {/* Desktop nav */}
          <nav aria-label="Navigasi utama" className="hidden items-center gap-0.5 md:flex">
            {navLinks.map((l) => {
              const active = pathname === l.href
              return (
                <Link
                  key={l.href}
                  href={l.href}
                  aria-current={active ? 'page' : undefined}
                  className={cn(
                    'rounded-full px-3.5 py-1.5 text-sm transition',
                    active
                      ? 'bg-st-surface-hi font-medium text-st-text'
                      : 'text-st-muted hover:bg-st-surface-hi hover:text-st-text',
                  )}
                >
                  {l.label}
                </Link>
              )
            })}
          </nav>

          <div className="flex items-center gap-1.5">
            <AnimatedThemeToggler
              theme={theme === 'dark' ? 'dark' : 'light'}
              onThemeChange={setTheme}
              className="inline-flex h-9 w-9 items-center justify-center rounded-full text-st-muted transition hover:bg-st-surface-hi hover:text-st-text [&_svg]:h-4.5 [&_svg]:w-4.5"
            />
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
              aria-expanded={open}
              aria-controls="mobile-nav"
              onClick={() => setOpen((v) => !v)}
              className="inline-flex h-9 w-9 items-center justify-center rounded-full text-st-muted transition hover:bg-st-surface-hi hover:text-st-text md:hidden"
            >
              {open ? <X className="h-5 w-5" /> : <Menu className="h-5 w-5" />}
            </button>
          </div>
        </div>

        {/* Tap-outside backdrop — closes the menu instead of swallowing the tap */}
        {open && (
          <div
            aria-hidden
            onClick={() => setOpen(false)}
            className="fixed inset-0 z-40 bg-transparent md:hidden"
          />
        )}

        {/* Mobile nav — absolute to the pill, above the header's stacking
            context so no safe-area offset can cover the menu items. */}
        {open && (
          <nav
            id="mobile-nav"
            aria-label="Navigasi mobile"
            className="absolute inset-x-0 top-full z-50 mt-2 max-h-[75dvh] overflow-y-auto rounded-2xl border border-st-stroke bg-(--st-surface)/90 p-2 pb-[max(0.5rem,env(safe-area-inset-bottom))] shadow-lg shadow-black/10 backdrop-blur-md md:hidden"
          >
            {navLinks.map((l) => {
              const active = pathname === l.href
              return (
                <Link
                  key={l.href}
                  href={l.href}
                  aria-current={active ? 'page' : undefined}
                  className={cn(
                    'block rounded-xl px-4 py-3 text-sm transition',
                    active
                      ? 'bg-st-surface-hi font-medium text-st-text'
                      : 'text-st-muted hover:bg-st-surface-hi hover:text-st-text',
                  )}
                >
                  {l.label}
                </Link>
              )
            })}
            <Link
              href={accessToken ? '/dashboard' : '/login'}
              className="mt-1 flex items-center justify-between rounded-xl bg-st-text px-4 py-3 text-sm font-medium text-st-bg"
            >
              {accessToken ? 'Buka Dashboard' : 'Masuk ke Sistem'}
              <ArrowRight className="h-4 w-4" />
            </Link>
          </nav>
        )}
      </div>
    </header>
  )
}
