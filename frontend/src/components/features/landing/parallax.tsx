'use client'

import { type ReactNode } from 'react'
import { useEffect, useRef } from 'react'

/**
 * Parallax — subtly decelerates a decorative layer as the page scrolls,
 * creating depth. rAF-throttled and disabled under prefers-reduced-motion.
 * Only safe on absolutely-positioned decorative layers; the layer is
 * translated relative to the section it belongs to, so it must not affect
 * document flow.
 */
export function Parallax({
  children,
  speed = 0.6,
  className,
}: {
  children: ReactNode
  speed?: number
  className?: string
}) {
  const ref = useRef<HTMLDivElement>(null)

  useEffect(() => {
    const el = ref.current
    if (!el) return
    if (window.matchMedia('(prefers-reduced-motion: reduce)').matches) return

    let frame = 0
    let active = false
    const update = () => {
      frame = 0
      if (!active) return
      const rect = el.getBoundingClientRect()
      const mid = rect.top + rect.height / 2 - window.innerHeight / 2
      const shift = Math.max(-72, Math.min(72, mid * (1 - speed)))
      el.style.transform = `translate3d(0, ${shift.toFixed(1)}px, 0)`
    }
    const onScroll = () => {
      if (!frame) frame = requestAnimationFrame(update)
    }
    window.addEventListener('scroll', onScroll, { passive: true })
    window.addEventListener('resize', onScroll)

    // Only do work while the layer is near the viewport — without this the
    // listener forces layout on every scroll frame for the whole page.
    const io = new IntersectionObserver(
      ([entry]) => {
        active = entry.isIntersecting
        if (active) update()
      },
      { rootMargin: '25% 0px' },
    )
    io.observe(el)

    return () => {
      io.disconnect()
      window.removeEventListener('scroll', onScroll)
      window.removeEventListener('resize', onScroll)
      if (frame) cancelAnimationFrame(frame)
    }
  }, [speed])

  return (
    <div ref={ref} className={className} aria-hidden>
      {children}
    </div>
  )
}