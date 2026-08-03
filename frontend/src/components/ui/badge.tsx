import * as React from 'react'
import { cn } from '@/lib/utils/cn'

type BadgeVariant = 'default' | 'primary' | 'success' | 'warning' | 'danger' | 'muted' | 'secondary' | 'outline'

const variantClasses: Record<BadgeVariant, string> = {
  default: 'bg-primary/15 text-primary border border-primary/25',
  primary: 'bg-primary/15 text-primary border border-primary/25',
  success: 'bg-success/15 text-success border border-success/25',
  warning: 'bg-warning/15 text-warning border border-warning/25',
  danger: 'bg-danger/15 text-danger-700 border border-danger/25',
  muted: 'bg-muted text-muted-foreground border border-border',
  secondary: 'bg-secondary/15 text-secondary border border-secondary/25',
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
