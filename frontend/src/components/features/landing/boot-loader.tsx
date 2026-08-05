'use client'

import { useEffect, useState } from 'react'

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
  const [hidden, setHidden] = useState<boolean>(() => {
    if (typeof window === 'undefined') return true
    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    return reduce || sessionStorage.getItem('st_booted') !== null
  })

  useEffect(() => {
    if (typeof window === 'undefined') return

    const reduce = window.matchMedia('(prefers-reduced-motion: reduce)').matches
    if (reduce || sessionStorage.getItem('st_booted')) {
      return
    }

    let raf = 0
    const hideRaf = requestAnimationFrame(() => setHidden(false))
    const start = performance.now()
    const tick = (now: number) => {
      const p = Math.min(1, (now - start) / DURATION)
      setCount(Math.round(p * 100))
      if (p < 1) {
        raf = requestAnimationFrame(tick)
      } else {
        window.setTimeout(() => {
          sessionStorage.setItem('st_booted', '1')
          setHidden(true)
        }, 340)
      }
    }
    raf = requestAnimationFrame(tick)
    return () => {
      cancelAnimationFrame(raf)
      cancelAnimationFrame(hideRaf)
    }
  }, [])

  if (hidden) return null

  const idx = Math.min(INIT_LINES.length - 1, Math.floor((count / 100) * INIT_LINES.length))

  return (
    <div
      className="fixed inset-0 z-[9999] flex flex-col justify-between px-6 py-8 md:px-10 md:py-10"
      style={{ background: 'var(--st-bg)' }}
    >
      <div className="flex items-center justify-between">
        <span className="font-mono text-[0.7rem] uppercase tracking-[0.3em] text-[var(--st-muted)]">
          SIMTAS://FILKOM
        </span>
        <span className="font-mono text-[0.7rem] uppercase tracking-[0.3em] text-[var(--st-muted)]">
          boot
        </span>
      </div>

      <div className="flex flex-col items-center justify-center gap-5">
        <p className="font-mono text-sm text-[var(--st-muted)]">
          <span className="text-[var(--st-accent-from)]">{'>'}</span> {INIT_LINES[idx]}…
        </p>
        <div className="font-display text-7xl tabular-nums text-[var(--st-text)] md:text-9xl">
          {String(count).padStart(3, '0')}
        </div>
      </div>

      <div className="h-[3px] w-full overflow-hidden rounded-full bg-[var(--st-stroke)]">
        <div
          className="accent-gradient h-full transition-[width] duration-75 ease-out"
          style={{ width: `${count}%`, boxShadow: '0 0 12px rgba(137,170,204,0.35)' }}
        />
      </div>
    </div>
  )
}
