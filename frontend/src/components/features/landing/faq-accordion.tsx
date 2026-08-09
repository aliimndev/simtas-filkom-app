'use client'

import { useState } from 'react'
import { ChevronDown } from 'lucide-react'
import { cn } from '@/lib/utils/cn'

export interface FaqItem {
  q: string
  a: string
}

type Variant = 'cards' | 'list'

const wrapperClasses: Record<Variant, string> = {
  cards: 'space-y-3',
  list: '',
}

const itemClasses: Record<Variant, string> = {
  cards: 'overflow-hidden rounded-2xl border border-st-stroke bg-st-surface',
  list: 'border-b border-st-stroke',
}

const buttonClasses: Record<Variant, string> = {
  cards: 'flex w-full items-center justify-between gap-4 px-5 py-4 text-left',
  list: 'flex w-full items-center justify-between gap-4 py-5 text-left',
}

const questionClasses: Record<Variant, string> = {
  cards: 'font-medium text-st-text',
  list: 'text-sm font-medium text-st-text md:text-base',
}

const answerClasses: Record<Variant, string> = {
  cards: 'border-t border-st-stroke px-5 py-4 text-sm leading-relaxed text-st-muted',
  list: 'pb-5 text-sm leading-relaxed text-st-muted',
}

/**
 * FaqAccordion — shared accordion used by the home page teaser
 * (variant="list", hairline rows) and the full /faq page (variant="cards",
 * rounded panels). One component owns the open/close state and the ARIA
 * wiring so both surfaces stay consistent.
 */
export function FaqAccordion({ items, variant = 'cards' }: { items: FaqItem[]; variant?: Variant }) {
  const [open, setOpen] = useState<number | null>(0)

  return (
    <div className={wrapperClasses[variant]}>
      {items.map((item, i) => {
        const isOpen = open === i
        return (
          <div key={item.q} className={itemClasses[variant]}>
            <button
              type="button"
              onClick={() => setOpen(isOpen ? null : i)}
              aria-expanded={isOpen}
              aria-controls={`faq-panel-${i}`}
              className={buttonClasses[variant]}
            >
              <span className={questionClasses[variant]}>{item.q}</span>
              <ChevronDown
                aria-hidden
                className={cn(
                  'h-4 w-4 shrink-0 text-st-muted transition-transform',
                  isOpen && 'rotate-180',
                )}
              />
            </button>
            {isOpen && (
              <p id={`faq-panel-${i}`} className={answerClasses[variant]}>
                {item.a}
              </p>
            )}
          </div>
        )
      })}
    </div>
  )
}
