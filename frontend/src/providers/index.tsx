'use client'

import { QueryProvider } from './query-provider'
import { ThemeProvider } from 'next-themes'

export function Providers({ children }: { children: React.ReactNode }) {
  return (
    <ThemeProvider attribute="class" defaultTheme="light" enableSystem>
      <QueryProvider>{children}</QueryProvider>
    </ThemeProvider>
  )
}