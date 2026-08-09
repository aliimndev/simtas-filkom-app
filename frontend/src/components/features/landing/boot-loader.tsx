'use client'

import { useLayoutEffect, useRef, useState } from 'react'

/**
 * BootLoader — a one-time, session-guarded "boot" overlay.
 *
 * The aesthetic risk: instead of a generic percentage counter, this reads
 * like a system initializing — terminal-style init lines + counter — which
 * is grounded in the computer-science subject of the faculty.
 *
 * It only runs once per browser session, is skipped entirely under
 * prefers-reduced-motion, and dismisses itself. Returning staff/students
 * never see it twice.
 */
const INIT_LINES = [
  'Memuat konfigurasi fakultas',
  'Menyambungkan repositori & arsip',
  'Menyiapkan alur Tugas Akhir',
  'Sistem siap',
]
const DURATION = 1700

export function BootLoader() {
  const [count, setCount] = useState(0)
  // SSR-safe: the initial state must be identical on server and client so
  // hydration matches (the server always renders null here). The overlay is
  // revealed inside useLayoutEffect — before the first client paint — once we
  // know the browser-only boot state.
  const [hidden, setHidden] = useState(true)
  // Once dismissed (timer or a tap), stop the animation loop entirely.
  const doneRef = useRef(false)

  useLayoutEffect(() => {
    if (typeof window === 'undefined') return

    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    if (reduce || sessionStorage.getItem('st_booted')) {
      return
    }

    let raf = 0
    // Defer the reveal to a callback so it runs after the layout effect (and
    // keeps the react-hooks set-state-in-effect lint rule happy).
    const showRaf = requestAnimationFrame(() => setHidden(false))
    const start = performance.now()
    const tick = (now: number) => {
      if (doneRef.current) return
      const p = Math.min(1, (now - start) / DURATION)
      setCount(Math.round(p * 100))
      if (p < 1) {
        raf = requestAnimationFrame(tick)
      } else {
        window.setTimeout(() => {
          sessionStorage.setItem('st_booted', '1')
          doneRef.current = true
          setHidden(true)
        }, 340)
      }
    }
    raf = requestAnimationFrame(tick)
    return () => {
      cancelAnimationFrame(raf)
      cancelAnimationFrame(showRaf)
    }
  }, [])

  if (hidden) return null

  const idx = Math.min(INIT_LINES.length - 1, Math.floor((count / 100) * INIT_LINES.length))

  // Dismiss the overlay immediately on tap so the boot sequence can never
  // swallow the visitor's first interaction (a common source of "clicked the
  // nav and nothing happened" on mobile).
  const skip = () => {
    try {
      sessionStorage.setItem('st_booted', '1')
    } catch {
      // private mode / storage unavailable — the overlay still hides
    }
    doneRef.current = true
    setHidden(true)
  }

  return (
    <div
      role="status"
      aria-label="Memuat SIMTAS FILKOM — ketuk untuk melewati"
      onClick={skip}
      className="fixed inset-0 z-9999 flex cursor-pointer flex-col justify-between bg-st-bg px-6 py-8 md:px-10 md:py-10"
    >
      <div className="flex items-center justify-between">
        <span className="font-mono text-[0.7rem] uppercase tracking-[0.3em] text-st-muted">
          SIMTAS://FILKOM
        </span>
        <span className="font-mono text-[0.7rem] uppercase tracking-[0.3em] text-st-muted">
          boot
        </span>
      </div>

      <div className="flex flex-col items-center justify-center gap-5">
        <p className="font-mono text-sm text-st-muted">
          <span className="text-(--st-accent-from)">{'>'}</span> {INIT_LINES[idx]}…
        </p>
        <div className="font-display text-7xl tabular-nums text-st-text md:text-9xl">
          {String(count).padStart(3, '0')}
        </div>
      </div>

      <div className="h-0.75 w-full overflow-hidden rounded-full bg-st-stroke">
        <div
          className="accent-gradient h-full transition-[width] duration-75 ease-out"
          style={{ width: `${count}%`, boxShadow: '0 0 12px rgba(137,170,204,0.35)' }}
        />
      </div>
    </div>
  )
}
