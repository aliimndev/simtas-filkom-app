import Link from 'next/link'
import type { ReactNode } from 'react'
import { cn } from '@/lib/utils/cn'

const baseClasses =
  'accent-ring inline-flex h-12 items-center justify-center gap-2 rounded-full px-7 text-sm font-medium transition'

type Variant = 'solid' | 'outline'

const variantClasses: Record<Variant, string> = {
  solid: 'bg-st-text text-st-bg hover:opacity-90',
  outline:
    'border border-st-stroke bg-st-surface text-st-text hover:border-(--st-accent-from)/40 hover:bg-st-surface-hi',
}

/**
 * LandingButton — the pill CTA shared by the hero and closing section.
 * Hash hrefs (in-page anchors) render as plain <a>; everything else uses
 * next/link so navigation stays client-side.
 */
export function LandingButton({
  href,
  variant = 'solid',
  className,
  children,
}: {
  href: string
  variant?: Variant
  className?: string
  children: ReactNode
}) {
  const classes = cn(baseClasses, variantClasses[variant], className)

  if (href.startsWith('#')) {
    return (
      <a href={href} className={classes}>
        {children}
      </a>
    )
  }

  return (
    <Link href={href} className={classes}>
      {children}
    </Link>
  )
}
