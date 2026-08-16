'use client'

import { useCallback, useLayoutEffect, useRef, type ReactNode } from 'react'
import Lenis from 'lenis'
import { cn } from '@/lib/utils/cn'

import './scroll-stack.css'

export function ScrollStackItem({ children }: { children: ReactNode }) {
  return <div className="scroll-stack-card">{children}</div>
}

export type ScrollStackProps = {
  /** ScrollStackItem components to stack. */
  children: ReactNode
  className?: string
  /** Gap (px) between cards before stacking. */
  itemDistance?: number
  /** Per-card scale delta (0.03 = each deeper card 3% smaller). */
  itemScale?: number
  /** Pixel gap between stacked cards. */
  itemStackDistance?: number
  /** Where stacking starts, as % of viewport height. */
  stackPosition?: string
  /** Where cards reach full scale, as % of viewport height. */
  scaleEndPosition?: string
  /** Scale of the deepest card. */
  baseScale?: number
  /** Unused — kept for API compatibility. */
  scaleDuration?: number
  /** Degrees of rotation applied per card while stacking. */
  rotationAmount?: number
  /** Blur (px) applied to cards deeper in the stack. */
  blurAmount?: number
  /** Use window scroll instead of a nested scroller. */
  useWindowScroll?: boolean
  onStackComplete?: () => void
}

export function ScrollStack({
  children,
  className = '',
  itemDistance = 100,
  itemScale = 0.03,
  itemStackDistance = 30,
  stackPosition = '20%',
  scaleEndPosition = '10%',
  baseScale = 0.85,
  scaleDuration = 0.5,
  rotationAmount = 0,
  blurAmount = 0,
  useWindowScroll = true,
  onStackComplete,
}: ScrollStackProps) {
  const scrollerRef = useRef<HTMLDivElement>(null)
  const stackCompletedRef = useRef(false)
  const animationFrameRef = useRef<number | null>(null)
  const lenisRef = useRef<Lenis | null>(null)
  const nativeCleanupRef = useRef<(() => void) | null>(null)
  const visibilityObserverRef = useRef<IntersectionObserver | null>(null)
  const isVisibleRef = useRef(true)
  const cardsRef = useRef<HTMLElement[]>([])
  const lastTransformsRef = useRef<Map<number, { translateY: number; scale: number; rotation: number; blur: number }>>(new Map())
  const isUpdatingRef = useRef(false)

  const calculateProgress = useCallback((scrollTop: number, start: number, end: number) => {
    if (scrollTop < start) return 0
    if (scrollTop > end) return 1
    return (scrollTop - start) / (end - start)
  }, [])

  const parsePercentage = useCallback((value: string, containerHeight: number) => {
    if (value.includes('%')) {
      return (parseFloat(value) / 100) * containerHeight
    }
    return parseFloat(value)
  }, [])

  const getScrollData = useCallback(() => {
    if (useWindowScroll) {
      return {
        scrollTop: window.scrollY,
        containerHeight: window.innerHeight,
      }
    }
    const scroller = scrollerRef.current
    return {
      scrollTop: scroller?.scrollTop ?? 0,
      containerHeight: scroller?.clientHeight ?? window.innerHeight,
    }
  }, [useWindowScroll])

  const getElementOffset = useCallback(
    (element: HTMLElement, scrollerTop?: number) => {
      if (useWindowScroll) {
        // offsetTop is layout-based and ignores transforms, so it does not
        // feed the card's own translate back into the next frame (no jitter).
        // Anchor it to the scroller, which is never transformed.
        const scroller = scrollerRef.current
        if (scroller) {
          // scrollerTop is precomputed once per update pass; reading
          // getBoundingClientRect per card forces layout every frame.
          const top =
            scrollerTop ?? scroller.getBoundingClientRect().top + window.scrollY
          return element.offsetTop + top
        }
      }
      return element.offsetTop
    },
    [useWindowScroll],
  )

  const updateCardTransforms = useCallback(() => {
    if (!cardsRef.current.length || isUpdatingRef.current) return
    // Skip all per-card layout math while the stack is off-screen: scroll
    // events from anywhere on the page would otherwise force layout here.
    if (!isVisibleRef.current) return
    isUpdatingRef.current = true

    const { scrollTop, containerHeight } = getScrollData()
    const stackPositionPx = parsePercentage(stackPosition, containerHeight)
    const scaleEndPositionPx = parsePercentage(scaleEndPosition, containerHeight)
    const endElement = useWindowScroll
      ? document.querySelector('.scroll-stack-end')
      : scrollerRef.current?.querySelector('.scroll-stack-end')
    const scrollerTop = scrollerRef.current
      ? scrollerRef.current.getBoundingClientRect().top + window.scrollY
      : undefined
    const endElementTop = endElement
      ? getElementOffset(endElement as HTMLElement, scrollerTop)
      : 0

    cardsRef.current.forEach((card, i) => {
      if (!card) return
      const cardTop = getElementOffset(card, scrollerTop)
      const triggerStart = cardTop - stackPositionPx - itemStackDistance * i
      const triggerEnd = cardTop - scaleEndPositionPx
      const pinStart = cardTop - stackPositionPx - itemStackDistance * i
      const pinEnd = endElementTop - containerHeight / 2
      const scaleProgress = calculateProgress(scrollTop, triggerStart, triggerEnd)
      const targetScale = baseScale + i * itemScale
      const scale = 1 - scaleProgress * (1 - targetScale)
      const rotation = rotationAmount ? i * rotationAmount * scaleProgress : 0

      let blur = 0
      if (blurAmount) {
        let topCardIndex = 0
        for (let j = 0; j < cardsRef.current.length; j++) {
          const jCardTop = getElementOffset(cardsRef.current[j], scrollerTop)
          const jTriggerStart = jCardTop - stackPositionPx - itemStackDistance * j
          if (scrollTop >= jTriggerStart) {
            topCardIndex = j
          }
        }
        if (i < topCardIndex) {
          const depthInStack = topCardIndex - i
          blur = Math.max(0, depthInStack * blurAmount)
        }
      }

      let translateY = 0
      const isPinned = scrollTop >= pinStart && scrollTop <= pinEnd
      if (isPinned) {
        translateY = scrollTop - cardTop + stackPositionPx + itemStackDistance * i
      } else if (scrollTop > pinEnd) {
        translateY = pinEnd - cardTop + stackPositionPx + itemStackDistance * i
      }

      const newTransform = {
        // Whole-pixel translate: prevents sub-pixel shimmer on 1x/2x DPI.
        translateY: Math.round(translateY),
        scale: Math.round(scale * 1000) / 1000,
        rotation: Math.round(rotation * 100) / 100,
        blur: Math.round(blur * 100) / 100,
      }

      const lastTransform = lastTransformsRef.current.get(i)
      const hasChanged =
        !lastTransform ||
        Math.abs(lastTransform.translateY - newTransform.translateY) > 0.1 ||
        Math.abs(lastTransform.scale - newTransform.scale) > 0.001 ||
        Math.abs(lastTransform.rotation - newTransform.rotation) > 0.1 ||
        Math.abs(lastTransform.blur - newTransform.blur) > 0.1

      if (hasChanged) {
        const transform = `translate3d(0, ${newTransform.translateY}px, 0) scale(${newTransform.scale}) rotate(${newTransform.rotation}deg)`
        const filter = newTransform.blur > 0 ? `blur(${newTransform.blur}px)` : ''
        card.style.transform = transform
        card.style.filter = filter
        lastTransformsRef.current.set(i, newTransform)
      }

      if (i === cardsRef.current.length - 1) {
        const isInView = scrollTop >= pinStart && scrollTop <= pinEnd
        if (isInView && !stackCompletedRef.current) {
          stackCompletedRef.current = true
          onStackComplete?.()
        } else if (!isInView && stackCompletedRef.current) {
          stackCompletedRef.current = false
        }
      }
    })

    isUpdatingRef.current = false
  }, [
    itemScale,
    itemStackDistance,
    stackPosition,
    scaleEndPosition,
    baseScale,
    rotationAmount,
    blurAmount,
    useWindowScroll,
    onStackComplete,
    calculateProgress,
    parsePercentage,
    getScrollData,
    getElementOffset,
  ])

  const handleScroll = useCallback(() => {
    updateCardTransforms()
  }, [updateCardTransforms])

  const setupLenis = useCallback(() => {
    // Lenis's syncTouch hijacks native touch scrolling and drives it from a
    // lerp + rAF loop, which stutters on mobile. Keep the smooth-wheel feel
    // only for fine-pointer (desktop) users without a reduced-motion
    // preference; touch devices use native scrolling with a passive,
    // rAF-throttled listener instead.
    const shouldUseLenis =
      window.matchMedia('(pointer: fine)').matches &&
      !window.matchMedia('(prefers-reduced-motion: reduce)').matches

    if (!shouldUseLenis) {
      let frame = 0
      const onScroll = () => {
        if (frame) return
        frame = requestAnimationFrame(() => {
          frame = 0
          updateCardTransforms()
        })
      }
      window.addEventListener('scroll', onScroll, { passive: true })
      nativeCleanupRef.current = () => {
        window.removeEventListener('scroll', onScroll)
        if (frame) cancelAnimationFrame(frame)
        frame = 0
      }
      return
    }

    const baseOptions = {
      duration: 1.2,
      easing: (t: number) => Math.min(1, 1.001 - Math.pow(2, -10 * t)),
      smoothWheel: true,
      touchMultiplier: 2,
      wheelMultiplier: 1,
      lerp: 0.1,
      syncTouch: true,
      syncTouchLerp: 0.075,
    }

    if (useWindowScroll) {
      const lenis = new Lenis({
        ...baseOptions,
        infinite: false,
      })
      lenis.on('scroll', handleScroll)
      const raf = (time: number) => {
        lenis.raf(time)
        animationFrameRef.current = requestAnimationFrame(raf)
      }
      animationFrameRef.current = requestAnimationFrame(raf)
      lenisRef.current = lenis
      return
    }

    const scroller = scrollerRef.current
    if (!scroller) return
    const lenis = new Lenis({
      ...baseOptions,
      wrapper: scroller,
      content: scroller.querySelector('.scroll-stack-inner') as HTMLElement,
      infinite: false,
    })
    lenis.on('scroll', handleScroll)
    const raf = (time: number) => {
      lenis.raf(time)
      animationFrameRef.current = requestAnimationFrame(raf)
    }
    animationFrameRef.current = requestAnimationFrame(raf)
    lenisRef.current = lenis
  }, [handleScroll, useWindowScroll])

  useLayoutEffect(() => {
    const scroller = scrollerRef.current
    if (!scroller) return

    const cards = Array.from(
      useWindowScroll
        ? document.querySelectorAll('.scroll-stack-card')
        : scroller.querySelectorAll('.scroll-stack-card'),
    ) as HTMLElement[]
    cardsRef.current = cards

    const transformsCache = lastTransformsRef.current
    cards.forEach((card, i) => {
      if (i < cards.length - 1) {
        card.style.marginBottom = `${itemDistance}px`
      }
      card.style.willChange = blurAmount > 0 ? 'transform, filter' : 'transform'
      card.style.transformOrigin = 'top center'
      card.style.backfaceVisibility = 'hidden'
      card.style.transform = 'translateZ(0)'
      card.style.webkitTransform = 'translateZ(0)'
      card.style.perspective = '1000px'
      card.style.webkitPerspective = '1000px'
    })

    setupLenis()
    updateCardTransforms()

    // Pause the stack's scroll work while it is away from the viewport.
    const visibilityObserver = new IntersectionObserver(
      ([entry]) => {
        isVisibleRef.current = entry.isIntersecting
        if (entry.isIntersecting) updateCardTransforms()
      },
      { rootMargin: '50% 0px' },
    )
    visibilityObserver.observe(scroller)
    visibilityObserverRef.current = visibilityObserver

    return () => {
      if (animationFrameRef.current !== null) {
        cancelAnimationFrame(animationFrameRef.current)
        animationFrameRef.current = null
      }
      visibilityObserverRef.current?.disconnect()
      visibilityObserverRef.current = null
      isVisibleRef.current = true
      nativeCleanupRef.current?.()
      nativeCleanupRef.current = null
      lenisRef.current?.destroy()
      lenisRef.current = null
      stackCompletedRef.current = false
      cardsRef.current = []
      transformsCache.clear()
      isUpdatingRef.current = false
    }
  }, [
    itemDistance,
    itemScale,
    itemStackDistance,
    stackPosition,
    scaleEndPosition,
    baseScale,
    scaleDuration,
    rotationAmount,
    blurAmount,
    useWindowScroll,
    onStackComplete,
    setupLenis,
    updateCardTransforms,
  ])

  return (
    <div ref={scrollerRef} className={cn('scroll-stack-scroller', className)}>
      <div className="scroll-stack-inner">
        {children}
        <div className="scroll-stack-end" aria-hidden />
      </div>
    </div>
  )
}
