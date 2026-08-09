'use client'

import { useRouter } from 'next/navigation'
import { useEffect } from 'react'
import { useAuthStore } from '@/lib/stores/auth-store'

/**
 * RedirectAuthenticated — sends already-authenticated visitors from the
 * public home straight to the dashboard once the persisted session has
 * hydrated. Renders nothing. Kept as its own client island so the landing
 * page itself stays a Server Component.
 */
export function RedirectAuthenticated() {
  const router = useRouter()
  const accessToken = useAuthStore((s) => s.accessToken)
  const isHydrated = useAuthStore((s) => s.isHydrated)

  useEffect(() => {
    if (isHydrated && accessToken) router.replace('/dashboard')
  }, [isHydrated, accessToken, router])

  return null
}
