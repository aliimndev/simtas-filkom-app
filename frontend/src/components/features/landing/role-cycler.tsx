'use client'

import { useEffect, useState } from 'react'

const ROLES = ['mahasiswa', 'dosen pembimbing', 'dosen penguji', 'kaprodi', 'admin fakultas']
const INTERVAL_MS = 2200

/**
 * RoleCycler — cycles the role word in the hero copy on a timed loop.
 * aria-live="polite" announces the change to assistive tech without
 * interrupting; reduced-motion is handled by the global CSS rule.
 */
export function RoleCycler() {
  const [i, setI] = useState(0)

  useEffect(() => {
    // Respect prefers-reduced-motion: keep the first role static instead of
    // cycling (matches the global CSS treatment of the animation).
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return
    const id = setInterval(() => setI((v) => (v + 1) % ROLES.length), INTERVAL_MS)
    return () => clearInterval(id)
  }, [])

  return (
    <span aria-live="polite" className="font-display italic text-st-text">
      <span key={i} className="animate-st-role inline-block">
        {ROLES[i]}
      </span>
    </span>
  )
}
