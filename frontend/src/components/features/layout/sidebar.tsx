'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import {
  ChevronDown,
  LogOut,
  Menu,
  UserRound,
  Search,
  Settings,
  HelpCircle,
  ChevronsUpDown,
} from 'lucide-react'
import { useState } from 'react'
import { navItemsForRoles, APP_NAME } from '@/constants/navigation'
import { roleLabel } from '@/constants/roles'
import { cn } from '@/lib/utils/cn'
import { useLogoutMutation } from '@/lib/hooks/use-auth'
import { useAuthStore } from '@/lib/stores/auth-store'
import { NotificationBell } from '@/components/features/notification-bell'
import { AnimatedThemeToggler } from '@/components/ui/animated-theme-toggler'
import { useTheme } from '@/providers/theme-provider'

/* ── Brand Mark — public editorial "sf" mark (mirrors landing nav/footer) ── */
function BrandMark() {
  return (
    <Link href="/dashboard" className="accent-ring relative flex h-9 w-9 shrink-0 items-center justify-center rounded-full">
      <span className="flex h-[calc(100%-3px)] w-[calc(100%-3px)] items-center justify-center rounded-full bg-st-surface">
        <span className="font-display text-[15px] italic leading-none text-st-text">sf</span>
      </span>
    </Link>
  )
}

/* ── Sidebar NavLink — editorial (public-style micro-labels + muted hover) ── */
function NavLink({ href, label, icon: Icon, onNavigate }: { href: string; label: string; icon: React.ElementType; onNavigate?: () => void }) {
  const pathname = usePathname()
  const active = pathname === href || (href !== '/dashboard' && pathname.startsWith(href))
  return (
    <Link
      href={href}
      onClick={onNavigate}
      className={cn(
        'flex items-center gap-2.5 rounded-full px-3 py-1.5 text-[13px] font-medium transition-colors duration-100',
        active
          ? 'bg-st-surface-hi text-st-text'
          : 'text-st-muted hover:bg-st-surface-hi hover:text-st-text',
      )}
    >
      <Icon className={cn('h-4 w-4 shrink-0', active ? 'text-(--st-accent-to)' : '')} />
      {label}
    </Link>
  )
}

/* ── Sidebar ── */
export function Sidebar({ onNavigate }: { onNavigate?: () => void }) {
  const user = useAuthStore((s) => s.user)
  const sections = navItemsForRoles(user ? [user.role] : [])

  return (
    <aside className="flex h-full w-60 flex-col border-r border-st-stroke bg-st-surface">
      {/* Brand header */}
      <div className="flex h-14 items-center gap-2 border-b border-st-stroke px-4">
        <BrandMark />
        <div className="min-w-0 flex-1">
          <p className="truncate font-display text-[15px] leading-none text-st-text">
            SIMTAS <span className="text-st-muted">FILKOM</span>
          </p>
        </div>
        <ChevronsUpDown className="h-3.5 w-3.5 shrink-0 text-st-muted" />
      </div>

      {/* Navigation */}
      <nav className="flex-1 overflow-y-auto px-3 py-4">
        <div className="space-y-6">
          {sections.map((section, i) => (
            <div key={i}>
              {section.title && (
                <p className="mb-2 px-3 font-mono text-[0.7rem] uppercase tracking-[0.25em] text-st-muted">
                  {section.title}
                </p>
              )}
              <div className="space-y-0.5">
                {section.items.map((item) => (
                  <NavLink key={item.href} href={item.href} label={item.label} icon={item.icon} onNavigate={onNavigate} />
                ))}
              </div>
            </div>
          ))}
        </div>

        {/* Bottom nav items */}
        <div className="mt-6 border-t border-st-stroke pt-4">
          <div className="space-y-0.5">
            <Link
              href="/profile"
              onClick={onNavigate}
              className="flex items-center gap-2.5 rounded-full px-3 py-1.5 text-[13px] font-medium text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text"
            >
              <Settings className="h-4 w-4" />
              Pengaturan
            </Link>
            <button
              type="button"
              className="flex w-full items-center gap-2.5 rounded-full px-3 py-1.5 text-[13px] font-medium text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text"
              aria-label="Bantuan"
            >
              <HelpCircle className="h-4 w-4" />
              Bantuan
            </button>
          </div>
        </div>
      </nav>

      {/* User section */}
      <div className="border-t border-st-stroke px-3 py-2">
        <Link
          href="/profile"
          onClick={onNavigate}
          className="flex items-center gap-2.5 rounded-full px-2.5 py-2 text-sm transition-colors hover:bg-st-surface-hi"
        >
          <div className="flex h-7 w-7 shrink-0 items-center justify-center rounded-full bg-st-surface-hi font-display text-[11px] italic text-st-text">
            {(user?.full_name ?? '?').slice(0, 1).toUpperCase()}
          </div>
          <div className="min-w-0 flex-1 leading-tight">
            <p className="truncate text-[13px] font-medium text-st-text">{user?.full_name ?? 'User'}</p>
            <p className="truncate text-[11px] text-st-muted">{roleLabel(user?.role)}</p>
          </div>
        </Link>
      </div>
    </aside>
  )
}


/* ── TopBar ── */
export function TopBar({ onOpenMenu }: { onOpenMenu?: () => void }) {
  const user = useAuthStore((s) => s.user)
  const logout = useLogoutMutation()
  const [menuOpen, setMenuOpen] = useState(false)
  const pathname = usePathname()
  const { theme, setTheme } = useTheme()

  const breadcrumb = buildBreadcrumb(pathname)

  return (
    <header className="flex h-12 items-center justify-between gap-3 border-b border-st-stroke bg-st-surface px-4 lg:px-5">
      <div className="flex min-w-0 items-center gap-2">
        {/* Mobile menu button */}
        {onOpenMenu && (
          <button
            type="button"
            onClick={onOpenMenu}
            className="rounded-full p-1.5 text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text lg:hidden"
            aria-label="Buka menu"
          >
            <Menu className="h-5 w-5" />
          </button>
        )}
        {/* Mobile app name */}
        <span className="truncate font-display text-[15px] text-st-text lg:hidden">{APP_NAME}</span>

        {/* Desktop breadcrumb */}
        <nav aria-label="Breadcrumb" className="hidden lg:block">
          <ol className="flex items-center gap-1 font-mono text-[0.7rem] uppercase tracking-[0.15em]">
            {breadcrumb.map((item, i) => (
              <li key={i} className="flex items-center gap-1">
                {i > 0 && <span className="text-st-muted/40" aria-hidden>/</span>}
                {i < breadcrumb.length - 1 ? (
                  <Link href={item.href} className="text-st-muted transition-colors hover:text-st-text">
                    {item.label}
                  </Link>
                ) : (
                  <span className="font-medium text-st-text">{item.label}</span>
                )}
              </li>
            ))}
          </ol>
        </nav>
      </div>

      {/* Right controls */}
      <div className="flex items-center gap-1">
        {/* Search */}
        <button
          type="button"
          className="hidden rounded-full p-1.5 text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text sm:block"
          aria-label="Cari"
        >
          <Search className="h-4.5 w-4.5" />
        </button>

        {/* Theme toggle — same animated switcher as the public navbar */}
        <AnimatedThemeToggler
          theme={theme === 'dark' ? 'dark' : 'light'}
          onThemeChange={setTheme}
          className="rounded-full p-1.5 text-st-muted transition-colors hover:bg-st-surface-hi hover:text-st-text [&_svg]:h-4.5 [&_svg]:w-4.5"
        />

        <NotificationBell />

        {/* User menu */}
        <div className="relative">
          <button
            type="button"
            onClick={() => setMenuOpen((v) => !v)}
            className="flex items-center gap-1.5 rounded-full px-1.5 py-1 text-sm font-medium text-st-text transition-colors hover:bg-st-surface-hi"
            aria-haspopup="menu"
            aria-expanded={menuOpen}
          >
            <span className="flex h-6 w-6 items-center justify-center rounded-full bg-st-surface-hi font-display text-[10px] italic text-st-text">
              {(user?.full_name ?? '?').slice(0, 1).toUpperCase()}
            </span>
            <ChevronDown className="h-3 w-3 text-st-muted" />
          </button>
          {menuOpen && (
            <div
              className="absolute right-0 top-full z-50 mt-1 w-48 overflow-hidden rounded-2xl border border-st-stroke bg-st-surface py-1 shadow-lg shadow-black/10"
              role="menu"
            >
              <div className="border-b border-st-stroke px-3 py-2">
                <p className="text-[13px] font-medium text-st-text">{user?.full_name ?? 'User'}</p>
                <p className="text-[11px] text-st-muted">{roleLabel(user?.role)}</p>
              </div>
              <Link
                href="/profile"
                className="flex items-center gap-2 px-3 py-2 text-[13px] text-st-text transition-colors hover:bg-st-surface-hi"
                onClick={() => setMenuOpen(false)}
                role="menuitem"
              >
                <UserRound className="h-3.5 w-3.5" /> Profil Saya
              </Link>
              <button
                type="button"
                onClick={() => {
                  setMenuOpen(false)
                  logout.mutate()
                }}
                className="flex w-full items-center gap-2 px-3 py-2 text-left text-[13px] text-danger-700 transition-colors hover:bg-danger/10"
                role="menuitem"
              >
                <LogOut className="h-3.5 w-3.5" /> Keluar
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}

/* ── Breadcrumb builder ── */
const ROUTE_LABELS: Record<string, string> = {
  dashboard: 'Dashboard',
  thesis: 'Tugas Akhir',
  theses: 'Daftar Tugas Akhir',
  supervision: 'Bimbingan',
  documents: 'Dokumen',
  seminars: 'Seminar',
  defenses: 'Sidang',
  archives: 'Arsip',
  profile: 'Profil',
  schedules: 'Jadwal',
  admin: 'Administrasi',
  users: 'Pengguna',
  'academic-years': 'Tahun Akademik',
  'audit-logs': 'Audit Log',
  'title-change-reviews': 'Review Perubahan Judul',
}

function buildBreadcrumb(pathname: string): { label: string; href: string }[] {
  const segments = pathname.split('/').filter(Boolean)
  const crumbs: { label: string; href: string }[] = []

  let path = ''
  for (const seg of segments) {
    path += `/${seg}`
    const label = ROUTE_LABELS[seg] ?? seg.replace(/-/g, ' ').replace(/\b\w/g, (c) => c.toUpperCase())
    crumbs.push({ label, href: path })
  }

  return crumbs.length > 0 ? crumbs : [{ label: 'Dashboard', href: '/dashboard' }]
}
