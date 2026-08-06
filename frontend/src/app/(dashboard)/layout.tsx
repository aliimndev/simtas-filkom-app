'use client'

import { Menu, X } from 'lucide-react'
import { useState } from 'react'
import { RequireAuth } from '@/components/features/require-auth'
import { Sidebar, TopBar } from '@/components/features/layout/sidebar'
import { Toaster } from '@/components/ui/toaster'

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <RequireAuth>
      <div className="flex min-h-screen">
        {/* Desktop sidebar */}
        <div className="hidden lg:block">
          <Sidebar />
        </div>

        {/* Mobile sidebar */}
        {mobileOpen && (
          <div className="fixed inset-0 z-50 lg:hidden">
            <div className="absolute inset-0 bg-black/50" onClick={() => setMobileOpen(false)} />
            <div className="absolute left-0 top-0 h-full w-64">
              <Sidebar onNavigate={() => setMobileOpen(false)} />
            </div>
            <button
              type="button"
              onClick={() => setMobileOpen(false)}
              className="absolute left-[16.25rem] top-3 rounded-full bg-background p-2 text-foreground shadow-lg"
              aria-label="Tutup menu"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
        )}

        <div className="flex min-w-0 flex-1 flex-col">
          <div className="lg:hidden">
            <div className="flex h-14 items-center justify-between border-b border-border bg-background px-4">
              <button
                type="button"
                onClick={() => setMobileOpen(true)}
                className="rounded-lg p-2 hover:bg-muted"
                aria-label="Buka menu"
              >
                <Menu className="h-5 w-5" />
              </button>
              <span className="text-sm font-bold">SIMTAS FILKOM</span>
              <span className="w-9" />
            </div>
          </div>
          <TopBar />
          <main className="flex-1 p-4 sm:p-6 lg:p-8">{children}</main>
        </div>
      </div>
      <Toaster />
    </RequireAuth>
  )
}
