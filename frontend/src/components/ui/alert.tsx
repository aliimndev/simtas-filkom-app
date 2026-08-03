import * as React from 'react'
import { AlertCircle, CheckCircle2, Info, XCircle } from 'lucide-react'
import { cn } from '@/lib/utils/cn'

type AlertVariant = 'info' | 'success' | 'warning' | 'danger'

const variantConfig: Record<AlertVariant, { container: string; icon: React.ReactNode }> = {
  info: {
    container: 'border-primary/25 bg-primary/10 text-primary',
    icon: <Info className="h-4 w-4 shrink-0" />,
  },
  success: {
    container: 'border-success/25 bg-success/10 text-success',
    icon: <CheckCircle2 className="h-4 w-4 shrink-0" />,
  },
  warning: {
    container: 'border-warning/25 bg-warning/10 text-warning',
    icon: <AlertCircle className="h-4 w-4 shrink-0" />,
  },
  danger: {
    container: 'border-danger/25 bg-danger/10 text-danger-700',
    icon: <XCircle className="h-4 w-4 shrink-0" />,
  },
}

export interface AlertProps extends React.HTMLAttributes<HTMLDivElement> {
  variant?: AlertVariant
}

export const Alert = React.forwardRef<HTMLDivElement, AlertProps>(
  ({ className, variant = 'info', children, ...props }, ref) => {
    const config = variantConfig[variant]
    return (
      <div
        ref={ref}
        role="alert"
        className={cn('flex items-start gap-3 rounded-lg border px-4 py-3 text-sm', config.container, className)}
        {...props}
      >
        {config.icon}
        <div className="min-w-0 flex-1">{children}</div>
      </div>
    )
  },
)
Alert.displayName = 'Alert'
