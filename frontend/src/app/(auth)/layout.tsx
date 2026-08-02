import type { Metadata } from 'next'

export const metadata: Metadata = {
  title: 'Autentikasi — SIMTAS FILKOM',
}

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return <>{children}</>
}
