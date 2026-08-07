import { create } from 'zustand'
import { persist } from 'zustand/middleware'
import type { User } from '@/types/auth'

interface AuthStore {
  user: User | null
  accessToken: string | null
  isHydrated: boolean
  setAuth: (user: User, accessToken: string) => void
  setUser: (user: User) => void
  clearAuth: () => void
  setHydrated: (v: boolean) => void
}

export const useAuthStore = create<AuthStore>()(
  persist(
    (set) => ({
      user: null,
      accessToken: null,
      isHydrated: false,
      setAuth: (user, accessToken) =>
        set({ user, accessToken, isHydrated: true }),
      setUser: (user) => set({ user }),
      clearAuth: () => set({ user: null, accessToken: null }),
      setHydrated: (v) => set({ isHydrated: v }),
    }),
    {
      name: 'simtas-auth',
      partialize: (state) => ({
        user: state.user,
        accessToken: state.accessToken,
      }),
      onRehydrateStorage: () => (state) => {
        state?.setHydrated(true)
      },
    },
  ),
)
