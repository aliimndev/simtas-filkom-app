'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ChevronDown, LogOut, Menu, UserRound } from 'lucide-react'
import { useState } from 'react'
import { navItemsForRoles, APP_NAME } from '@/constants/navigation'
import { roleLabel } from '@/constants/roles'
import { cn } from '@/lib/utils/cn'
import { useLogoutMutation } from '@/lib/hooks/use-auth'
import { useAuthStore } from '@/lib/stores/auth-store'
import { NotificationBell } from '@/components/features/notification-bell'

function NavLink({ href, label, icon: Icon, onNavigate }: { href: string; label: string; icon: React.ElementType; onNavigate?: () => void }) {
  const pathname = usePathname()
  const active = pathname === href || (href !== '/dashboard' && pathname.startsWith(href))
  return (
    <Link
      href={href}
      onClick={onNavigate}
      className={cn(
        'flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors duration-150',
        active
          ? 'bg-primary-50 text-primary-800'
          : 'text-muted-foreground hover:bg-surface-hi hover:text-foreground',
      )}
    >
      <Icon className="h-[18px] w-[18px] shrink-0" />
      {label}
    </Link>
  )
}

function BrandMark() {
  return (
    <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-md bg-primary text-sm font-bold text-primary-foreground">
      S
    </div>
  )
}

export function Sidebar({ onNavigate }: { onNavigate?: () => void }) {
  const { user } = useAuthStore()
  const sections = navItemsForRoles(user ? [user.role] : [])

  return (
    <aside className="flex h-full w-64 flex-col border-r border-border bg-card">
      <div className="flex h-14 items-center gap-2.5 border-b border-border px-5">
        <BrandMark />
        <div className="leading-tight">
          <p className="text-sm font-semibold tracking-tight text-foreground">{APP_NAME}</p>
          <p className="text-[11px] text-muted-foreground">Fakultas Ilmu Komputer Unida</p>
        </div>
      </div>

      <nav className="flex-1 space-y-6 overflow-y-auto px-3 py-4">
        {sections.map((section, i) => (
          <div key={i}>
            {section.title && (
              <p className="mb-1.5 px-3 text-[11px] font-medium uppercase tracking-wide text-muted-foreground">
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
      </nav>

      <div className="border-t border-border p-3">
        <Link
          href="/profile"
          onClick={onNavigate}
          className="flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors hover:bg-surface-hi"
        >
          <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-primary-50 text-xs font-semibold text-primary-800">
            {(user?.full_name ?? '?').slice(0, 1).toUpperCase()}
          </div>
          <div className="min-w-0 flex-1 leading-tight">
            <p className="truncate text-sm font-medium text-foreground">{user?.full_name ?? 'User'}</p>
            <p className="truncate text-[11px] text-muted-foreground">{roleLabel(user?.role)}</p>
          </div>
        </Link>
      </div>
    </aside>
  )
}

export function TopBar({ onOpenMenu }: { onOpenMenu?: () => void }) {
  const { user } = useAuthStore()
  const logout = useLogoutMutation()
  const [menuOpen, setMenuOpen] = useState(false)

  return (
    <header className="flex h-12 items-center justify-between gap-3 border-b border-border bg-card px-4 lg:px-6">
      <div className="flex min-w-0 items-center gap-2">
        {onOpenMenu && (
          <button
            type="button"
            onClick={onOpenMenu}
            className="rounded-md p-2 text-muted-foreground transition-colors hover:bg-surface-hi hover:text-foreground lg:hidden"
            aria-label="Buka menu"
          >
            <Menu className="h-5 w-5" />
          </button>
        )}
        <span className="truncate text-sm font-semibold text-foreground lg:hidden">{APP_NAME}</span>
      </div>

      <div className="flex items-center gap-1">
        <NotificationBell />
        <div className="relative">
          <button
            type="button"
            onClick={() => setMenuOpen((v) => !v)}
            className="flex items-center gap-2 rounded-md px-2 py-1.5 text-sm font-medium text-foreground transition-colors hover:bg-surface-hi"
            aria-haspopup="menu"
            aria-expanded={menuOpen}
          >
            <span className="flex h-7 w-7 items-center justify-center rounded-full bg-primary-50 text-xs font-semibold text-primary-800">
              {(user?.full_name ?? '?').slice(0, 1).toUpperCase()}
            </span>
            <span className="hidden sm:inline">{user?.full_name?.split(' ')[0] ?? 'Menu'}</span>
            <ChevronDown className="h-3.5 w-3.5 text-muted-foreground" />
          </button>
          {menuOpen && (
            <div
              className="absolute right-0 top-full z-50 mt-1.5 w-48 overflow-hidden rounded-md border border-border bg-card py-1"
              role="menu"
            >
              <Link
                href="/profile"
                className="flex items-center gap-2 px-4 py-2 text-sm text-foreground transition-colors hover:bg-surface-hi"
                onClick={() => setMenuOpen(false)}
                role="menuitem"
              >
                <UserRound className="h-4 w-4" /> Profil Saya
              </Link>
              <button
                type="button"
                onClick={() => {
                  setMenuOpen(false)
                  logout.mutate()
                }}
                className="flex w-full items-center gap-2 px-4 py-2 text-left text-sm text-danger-700 transition-colors hover:bg-danger/10"
                role="menuitem"
              >
                <LogOut className="h-4 w-4" /> Keluar
              </button>
            </div>
          )}
        </div>
      </div>
    </header>
  )
}
