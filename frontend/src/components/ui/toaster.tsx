'use client'

import { CheckCircle2, Info, X, XCircle } from 'lucide-react'
import { useEffect, useState } from 'react'
import { cn } from '@/lib/utils/cn'

export type ToastVariant = 'success' | 'error' | 'info'

export interface ToastOptions {
  title: string
  description?: string
  variant?: ToastVariant
}

interface ToastItem extends Omit<ToastOptions, 'variant'> {
  id: number
  variant: ToastVariant
}

type Listener = (toast: ToastItem) => void

let nextId = 0
const listeners = new Set<Listener>()

/** Panggil dari mana saja (di luar React sekalipun) untuk menampilkan toast. */
export function toast(options: ToastOptions) {
  const item: ToastItem = { id: ++nextId, variant: 'info', ...options }
  listeners.forEach((l) => l(item))
}

const variantConfig: Record<ToastVariant, { icon: React.ReactNode; iconClass: string; ringClass: string }> = {
  success: {
    icon: <CheckCircle2 className="h-5 w-5" />,
    iconClass: 'text-success',
    ringClass: 'border-success/25',
  },
  error: {
    icon: <XCircle className="h-5 w-5" />,
    iconClass: 'text-danger-700',
    ringClass: 'border-danger/25',
  },
  info: {
    icon: <Info className="h-5 w-5" />,
    iconClass: 'text-primary',
    ringClass: 'border-primary/25',
  },
}

const DURATION_MS = 4500

/**
 * Toast stack, mount sekali di layout dashboard (`<Toaster />`).
 * Auto-dismiss setelah beberapa detik; bisa ditutup manual.
 */
export function Toaster() {
  const [toasts, setToasts] = useState<ToastItem[]>([])

  useEffect(() => {
    const listener: Listener = (t) => {
      setToasts((prev) => [...prev.slice(-3), t])
      window.setTimeout(() => {
        setToasts((prev) => prev.filter((x) => x.id !== t.id))
      }, DURATION_MS)
    }
    listeners.add(listener)
    return () => {
      listeners.delete(listener)
    }
  }, [])

  return (
    <div
      aria-live="polite"
      className="pointer-events-none fixed bottom-4 right-4 z-[80] flex w-[calc(100vw-2rem)] max-w-sm flex-col gap-2"
    >
      {toasts.map((t) => {
        const cfg = variantConfig[t.variant]
        return (
          <div
            key={t.id}
            className={cn(
              'toast-in pointer-events-auto flex items-start gap-3 rounded-xl border bg-background px-4 py-3 shadow-lg',
              cfg.ringClass,
            )}
          >
            <span className={cn('mt-0.5 shrink-0', cfg.iconClass)}>{cfg.icon}</span>
            <div className="min-w-0 flex-1">
              <p className="text-sm font-medium text-foreground">{t.title}</p>
              {t.description && <p className="mt-0.5 text-xs leading-relaxed text-muted-foreground">{t.description}</p>}
            </div>
            <button
              type="button"
              onClick={() => setToasts((prev) => prev.filter((x) => x.id !== t.id))}
              aria-label="Tutup notifikasi"
              className="shrink-0 rounded-md p-0.5 text-muted-foreground/70 transition-colors hover:bg-muted hover:text-foreground"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </div>
        )
      })}
    </div>
  )
}
