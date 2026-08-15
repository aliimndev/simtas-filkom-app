import { type ReactNode } from 'react'

/**
 * DashboardHeroBand — atmospheric header band for the dashboard home pages.
 * Wraps greeting + stat cards with a dark surface, subtle grid, blur orbs,
 * and a feathered bottom edge (gradient fade into the canvas).
 */
export function DashboardHeroBand({ children }: { children: ReactNode }) {
  return (
    <div className="relative -mx-4 -mt-4 overflow-hidden sm:-mx-6 sm:-mt-6 lg:-mx-8 lg:-mt-8">
      {/* Atmospheric background */}
      <div className="absolute inset-0 bg-st-bg" />
      {/* Subtle grid pattern — reuse landing-grid-bg */}
      <div className="absolute inset-0 landing-grid-bg opacity-[0.04]" aria-hidden="true" />
      {/* Blur orbs */}
      <div className="absolute -right-20 -top-20 h-64 w-64 rounded-full bg-(--st-accent-from)/8 blur-[100px]" aria-hidden="true" />
      <div className="absolute -bottom-10 -left-10 h-48 w-48 rounded-full bg-(--st-accent-to)/6 blur-[80px]" aria-hidden="true" />

      {/* Content */}
      <div className="relative z-10 px-4 pt-4 pb-8 sm:px-6 sm:pt-6 sm:pb-10 lg:px-8 lg:pt-8 lg:pb-12">
        {children}
      </div>

      {/* Feathered bottom edge */}
      <div
        className="absolute bottom-0 left-0 right-0 h-12"
        style={{
          background: `linear-gradient(to bottom, transparent, var(--st-bg))`,
        }}
        aria-hidden="true"
      />
    </div>
  )
}
