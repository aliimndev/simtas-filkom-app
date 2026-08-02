'use client'

import { usePathname, useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { Spinner } from '@/components/ui/spinner'
import { useAuthStore } from '@/lib/stores/auth-store'

/**
 * Guard rute: mewajibkan user sudah login. Jika belum → redirect ke /login.
 * Opsional `roles`: membatasi akses sesuai peran (403 / redirect).
 */
export function RequireAuth({
  children,
  roles,
  redirectTo = '/login',
}: {
  children: React.ReactNode
  roles?: string[]
  redirectTo?: string
}) {
  const router = useRouter()
  const pathname = usePathname()
  const { accessToken, user, isHydrated } = useAuthStore()

  useEffect(() => {
    if (!isHydrated) return
    if (!accessToken) {
      router.replace(`${redirectTo}?next=${encodeURIComponent(pathname)}`)
      return
    }
    if (roles && user && !roles.includes(user.role)) {
      router.replace('/dashboard')
    }
  }, [isHydrated, accessToken, user, roles, router, pathname, redirectTo])

  if (!isHydrated || !accessToken) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner label="Memeriksa sesi…" />
      </div>
    )
  }

  if (roles && user && !roles.includes(user.role)) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <Spinner label="Mengalihkan…" />
      </div>
    )
  }

  return <>{children}</>
}
