'use client'

import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useRouter, useSearchParams } from 'next/navigation'
import { useEffect, useMemo } from 'react'
import { authApi } from '@/lib/api/auth-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import type { LoginRequest, ForgotPasswordRequest, ResetPasswordRequest } from '@/types/auth'

const meKey = ['auth', 'me'] as const

export function useCurrentUser() {
  const { accessToken, setUser, clearAuth } = useAuthStore()
  const query = useQuery({
    queryKey: meKey,
    queryFn: authApi.getMe,
    enabled: Boolean(accessToken),
    staleTime: 5 * 60 * 1000,
  })

  useEffect(() => {
    if (query.data) setUser(query.data)
  }, [query.data, setUser])

  useEffect(() => {
    if (query.isError) clearAuth()
  }, [query.isError, clearAuth])

  return query
}

export function useLoginMutation() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const queryClient = useQueryClient()
  const { setAuth } = useAuthStore()

  const nextPath = useMemo(() => {
    const next = searchParams.get('next')
    if (!next || !next.startsWith('/') || next.startsWith('//')) return null
    return next
  }, [searchParams])

  return useMutation({
    mutationFn: (data: LoginRequest) => authApi.login(data),
    onSuccess: (res) => {
      setAuth(res.user, res.access_token, res.refresh_token)
      queryClient.invalidateQueries({ queryKey: meKey })
      if (res.user.must_change_password) {
        router.push('/profile?force=change-password')
      } else {
        router.push(nextPath ?? '/dashboard')
      }
    },
  })
}

export function useLogoutMutation() {
  const router = useRouter()
  const queryClient = useQueryClient()
  const { clearAuth, accessToken } = useAuthStore()

  return useMutation({
    mutationFn: async () => {
      try {
        if (accessToken) await authApi.logout()
      } finally {
        clearAuth()
        queryClient.clear()
        router.push('/login')
      }
    },
  })
}

export function useForgotPasswordMutation() {
  return useMutation({
    mutationFn: (data: ForgotPasswordRequest) => authApi.forgotPassword(data),
  })
}

export function useResetPasswordMutation() {
  return useMutation({
    mutationFn: (data: ResetPasswordRequest) => authApi.resetPassword(data),
  })
}
