import * as React from 'react'
import { Loader2 } from 'lucide-react'
import { cn } from '@/lib/utils/cn'

type Variant = 'primary' | 'secondary' | 'danger' | 'success' | 'ghost' | 'outline' | 'default' | 'destructive' | 'icon' | 'link'
type Size = 'sm' | 'md' | 'lg' | 'xs' | 'icon' | 'icon-sm'

const variantClasses: Record<Variant, string> = {
  primary: 'bg-primary text-primary-foreground hover:bg-primary-700 focus-visible:ring-primary',
  default: 'bg-primary text-primary-foreground hover:bg-primary-700 focus-visible:ring-primary',
  secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary-700 focus-visible:ring-secondary',
  danger: 'bg-danger text-danger-foreground hover:bg-danger-700 focus-visible:ring-danger',
  destructive: 'bg-danger text-danger-foreground hover:bg-danger-700 focus-visible:ring-danger',
  success: 'bg-success text-success-foreground hover:bg-success/90 focus-visible:ring-success',
  ghost: 'bg-transparent hover:bg-muted text-foreground focus-visible:ring-ring',
  icon: 'bg-transparent text-foreground hover:bg-muted focus-visible:ring-ring',
  link: 'bg-transparent p-0 text-primary underline-offset-4 hover:underline focus-visible:ring-primary',
  outline:
    'border border-border bg-transparent hover:bg-muted text-foreground focus-visible:ring-ring',
}

const sizeClasses: Record<Size, string> = {
  xs: 'h-7 px-2 text-xs rounded-md gap-1',
  sm: 'min-h-10 sm:min-h-0 h-8 sm:h-9 px-3 text-xs rounded-md gap-1.5',
  md: 'h-10 sm:h-11 px-4 text-sm rounded-md gap-2',
  lg: 'h-12 px-6 text-base rounded-md gap-2',
  icon: 'h-8 w-8 px-0 rounded-md',
  'icon-sm': 'h-7 w-7 px-0 rounded-md',
}

export interface ButtonProps extends React.ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant
  size?: Size
  loading?: boolean
  fullWidth?: boolean
  /** Render sebagai elemen anak (mis. Link) sambil tetap memakai styling tombol. */
  asChild?: boolean
}

const buttonBaseClasses =
  'inline-flex items-center justify-center font-medium transition-colors duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 disabled:pointer-events-none disabled:opacity-50'

export const Button = React.forwardRef<HTMLButtonElement, ButtonProps>(
  (
    { className, variant = 'primary', size = 'md', loading = false, fullWidth = false, disabled, asChild = false, children, ...props },
    ref,
  ) => {
    const classes = cn(
      buttonBaseClasses,
      variantClasses[variant],
      sizeClasses[size],
      fullWidth && 'w-full',
      className,
    )

    if (asChild && React.isValidElement(children)) {
      const child = children as React.ReactElement<React.HTMLAttributes<HTMLElement>>
      return React.cloneElement(child, {
        className: cn(classes, child.props.className),
      })
    }

    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={classes}
        {...props}
      >
        {loading && <Loader2 className="h-4 w-4 animate-spin" />}
        {children}
      </button>
    )
  },
)
Button.displayName = 'Button'
