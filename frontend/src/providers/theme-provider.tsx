'use client'

import { createContext, useCallback, useContext, useEffect, useMemo, useState } from 'react'

export type Theme = 'light' | 'dark'

interface ThemeContextValue {
  theme: Theme
  setTheme: (theme: Theme) => void
}

const THEME_STORAGE_KEY = 'theme'

const ThemeContext = createContext<ThemeContextValue>({
  theme: 'light',
  setTheme: () => {},
})

/**
 * Client-only: read the persisted theme. On the server (and before any stored
 * value) the institutional default is light.
 */
function getStoredTheme(): Theme {
  if (typeof window === 'undefined') return 'light'
  try {
    return localStorage.getItem(THEME_STORAGE_KEY) === 'dark' ? 'dark' : 'light'
  } catch {
    return 'light'
  }
}

/**
 * Minimal theme provider replacing next-themes.
 *
 * Why custom: next-themes 0.4.6 (latest) unconditionally renders an inline
 * <script> element, which trips React 19's "Encountered a script tag" warning
 * in dev and, because it applies the stored theme before hydration, causes a
 * hydration mismatch in any theme-dependent UI. The pre-hydration script is
 * instead injected via next/script (beforeInteractive) in the root layout, so
 * FOUC protection is preserved without either error.
 */
export function ThemeProvider({ children }: { children: React.ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(getStoredTheme)

  // Keep <html> in sync after mount (the init script already applied the class
  // pre-hydration; this is idempotent and covers toggles from this tab).
  // color-scheme follows the `.dark` class via the CSS rules in globals.css.
  useEffect(() => {
    document.documentElement.classList.toggle('dark', theme === 'dark')
  }, [theme])

  // Cross-tab sync: a theme change in another tab follows here.
  useEffect(() => {
    const onStorage = (e: StorageEvent) => {
      if (e.key !== THEME_STORAGE_KEY) return
      if (e.newValue === 'dark' || e.newValue === 'light') setThemeState(e.newValue)
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [])

  const setTheme = useCallback((next: Theme) => {
    setThemeState(next)
    try {
      localStorage.setItem(THEME_STORAGE_KEY, next)
    } catch {
      // private mode / storage unavailable — the class still flips via effect
    }
  }, [])

  const value = useMemo(() => ({ theme, setTheme }), [theme, setTheme])

  return <ThemeContext.Provider value={value}>{children}</ThemeContext.Provider>
}

export function useTheme() {
  return useContext(ThemeContext)
}
