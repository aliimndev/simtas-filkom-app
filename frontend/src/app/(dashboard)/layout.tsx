'use client'

import { useState } from 'react'
import { X } from 'lucide-react'
import { RequireAuth } from '@/components/features/require-auth'
import { ErrorBoundary } from '@/components/features/error-boundary'
import { Sidebar, TopBar } from '@/components/features/layout/sidebar'
import { Toaster } from '@/components/ui/toaster'

/*
THESIS: a calm academic task console — the thesis workflow is the story, not the chrome.
OWN-WORLD: white cards on a neutral canvas, hairline borders, small radii (8px/6px), one sans
  family (Inter), a single restrained FILKOM teal accent, 8px spacing grid; no serif/mono
  costume, no emoji, no glow, no shadow lifts inside the dashboard surface.
STORY: the visitor scans their status, pending actions, and schedule in seconds and acts.
FIRST VIEWPORT: white sidebar + slim top bar (bell + user menu); content opens with a compact
  greeting header, then a status/progress band, pending actions, schedule, and activity.
FORM: brief-pinned direction (user brief), no concept roll.
FINISH: unreviewed and undocumented is unfinished; this build ends with the finish review, the verdict, and DESIGN.md.
*/

export default function DashboardLayout({ children }: { children: React.ReactNode }) {
  const [mobileOpen, setMobileOpen] = useState(false)

  return (
    <RequireAuth>
      <ErrorBoundary>
      <div className="flex min-h-screen bg-st-bg">
        {/* Desktop sidebar */}
        <div className="hidden lg:block">
          <Sidebar />
        </div>

        {/* Mobile sidebar */}
        {mobileOpen && (
          <div className="fixed inset-0 z-50 lg:hidden">
            <div aria-hidden className="absolute inset-0 bg-black/40" onClick={() => setMobileOpen(false)} />
            <div className="absolute left-0 top-0 h-full w-60">
              <Sidebar onNavigate={() => setMobileOpen(false)} />
            </div>
            <button
              type="button"
              onClick={() => setMobileOpen(false)}
              className="absolute left-61 top-3 rounded-full bg-st-surface p-2 text-st-text"
              aria-label="Tutup menu"
            >
              <X className="h-5 w-5" />
            </button>
          </div>
        )}

        <div className="flex min-w-0 flex-1 flex-col">
          <TopBar onOpenMenu={() => setMobileOpen(true)} />
          <main className="mx-auto w-full max-w-350 flex-1 p-4 sm:p-6 lg:p-8">{children}</main>
        </div>
      </div>
      <Toaster />
      </ErrorBoundary>
    </RequireAuth>
  )
}
