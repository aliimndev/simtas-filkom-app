import * as React from 'react'
import { cn } from '@/lib/utils/cn'

type BadgeVariant = 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'muted' | 'secondary' | 'outline'

const variantClasses: Record<BadgeVariant, string> = {
  default: 'bg-primary-50 text-primary border border-primary/20',
  primary: 'bg-primary-50 text-primary border border-primary/20',
  success: 'bg-success-50 text-success border border-success/20',
  warning: 'bg-warning-50 text-warning border border-warning/20',
  danger: 'bg-danger-50 text-danger-700 border border-danger/20',
  muted: 'bg-st-surface-hi text-st-muted border border-st-stroke',
  secondary: 'bg-secondary-50 text-secondary border border-secondary/20',
  outline: 'border border-st-stroke text-st-text',
}

export interface BadgeProps extends React.HTMLAttributes<HTMLSpanElement> {
  variant?: BadgeVariant
}

export const Badge = React.forwardRef<HTMLSpanElement, BadgeProps>(
  ({ className, variant = 'default', ...props }, ref) => (
    <span
      ref={ref}
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2.5 py-0.5 text-xs font-medium',
        variantClasses[variant],
        className,
      )}
      {...props}
    />
  ),
)
Badge.displayName = 'Badge'
