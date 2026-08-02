'use client'

import Link from 'next/link'
import { usePathname } from 'next/navigation'
import { ChevronDown, LogOut, Settings, UserRound } from 'lucide-react'
import { useState } from 'react'
import { navItemsForRoles, APP_NAME } from '@/constants/navigation'
import { roleLabel } from '@/constants/roles'
import { cn } from '@/lib/utils/cn'
import { useLogoutMutation } from '@/lib/hooks/use-auth'
import { useAuthStore } from '@/lib/stores/auth-store'

function NavLink({ href, label, icon: Icon, onNavigate }: { href: string; label: string; icon: React.ElementType; onNavigate?: () => void }) {
  const pathname = usePathname()
  const active = pathname === href || (href !== '/dashboard' && pathname.startsWith(href))
  return (
    <Link
      href={href}
      onClick={onNavigate}
      className={cn(
        'flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm font-medium transition-colors',
        active
          ? 'bg-sidebar-accent text-white shadow-sm'
          : 'text-sidebar-foreground/80 hover:bg-sidebar-accent/50 hover:text-white',
      )}
    >
      <Icon className="h-4.5 w-4.5 shrink-0" />
      {label}
    </Link>
  )
}

export function Sidebar({ onNavigate }: { onNavigate?: () => void }) {
  const { user } = useAuthStore()
  const sections = navItemsForRoles(user ? [user.role] : [])

  return (
    <aside className="flex h-full w-64 flex-col bg-sidebar text-sidebar-foreground">
      <div className="flex h-14 items-center gap-2.5 border-b border-sidebar-border px-5">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-sm font-black text-primary-foreground">
          SF
        </div>
        <div className="leading-tight">
          <p className="text-sm font-bold">{APP_NAME}</p>
          <p className="text-[10px] text-sidebar-foreground/60">FILKOM Unida</p>
        </div>
      </div>

      <nav className="flex-1 space-y-5 overflow-y-auto px-3 py-4">
        {sections.map((section, i) => (
          <div key={i}>
            {section.title && (
              <p className="mb-1.5 px-3 text-[10px] font-semibold uppercase tracking-widest text-sidebar-foreground/40">
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

      <div className="border-t border-sidebar-border p-3">
        <Link
          href="/profile"
          onClick={onNavigate}
          className="flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm transition-colors hover:bg-sidebar-accent/50"
        >
          <div className="flex h-8 w-8 items-center justify-center rounded-full bg-primary/30 text-xs font-bold text-white">
            {(user?.full_name ?? '?').slice(0, 1).toUpperCase()}
          </div>
          <div className="min-w-0 flex-1 leading-tight">
            <p className="truncate text-sm font-medium text-white">{user?.full_name ?? 'User'}</p>
            <p className="truncate text-[11px] text-sidebar-foreground/60">{roleLabel(user?.role)}</p>
          </div>
          <ChevronDown className="h-4 w-4 text-sidebar-foreground/50" />
        </Link>
      </div>
    </aside>
  )
}

export function TopBar() {
  const { user } = useAuthStore()
  const logout = useLogoutMutation()
  const [menuOpen, setMenuOpen] = useState(false)

  return (
    <header className="flex h-14 items-center justify-between border-b border-border bg-background px-6">
      <div>
        <p className="text-sm font-medium">Selamat datang,</p>
        <p className="text-xs text-muted-foreground">{roleLabel(user?.role)}</p>
      </div>

      <div className="relative">
        <button
          type="button"
          onClick={() => setMenuOpen((v) => !v)}
          className="flex items-center gap-2 rounded-lg px-3 py-1.5 text-sm font-medium transition-colors hover:bg-muted"
        >
          <Settings className="h-4 w-4 text-muted-foreground" />
          <span className="hidden sm:inline">Menu</span>
        </button>
        {menuOpen && (
          <div
            className="absolute right-0 top-full z-50 mt-2 w-48 overflow-hidden rounded-xl border border-border bg-background shadow-xl"
            onMouseLeave={() => setMenuOpen(false)}
          >
            <Link
              href="/profile"
              className="flex items-center gap-2 px-4 py-2.5 text-sm transition-colors hover:bg-muted"
              onClick={() => setMenuOpen(false)}
            >
              <UserRound className="h-4 w-4" /> Profil Saya
            </Link>
            <button
              type="button"
              onClick={() => {
                setMenuOpen(false)
                logout.mutate()
              }}
              className="flex w-full items-center gap-2 px-4 py-2.5 text-left text-sm text-danger-700 transition-colors hover:bg-danger-50"
            >
              <LogOut className="h-4 w-4" /> Keluar
            </button>
          </div>
        )}
      </div>
    </header>
  )
}
