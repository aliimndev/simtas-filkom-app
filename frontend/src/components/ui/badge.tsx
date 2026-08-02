import * as React from 'react'
import { cn } from '@/lib/utils/cn'

type BadgeVariant = 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'muted' | 'secondary' | 'outline'

const variantClasses: Record<BadgeVariant, string> = {
  default: 'bg-primary text-primary-foreground',
  primary: 'bg-primary-100 text-primary-800 border border-primary-200',
  success: 'bg-success-100 text-success border border-success/30',
  warning: 'bg-warning-100 text-warning border border-warning/30',
  danger: 'bg-danger-100 text-danger-700 border border-danger/30',
  muted: 'bg-muted text-muted-foreground border border-border',
  secondary: 'bg-secondary-100 text-secondary-foreground border border-secondary/30',
  outline: 'border border-border text-foreground',
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
