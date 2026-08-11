'use client'

import { NavbarSection } from '@/components/features/landing/navbar'
import { FooterSection } from '@/components/features/landing/footer'

const NAV_LINKS = [
  { href: '/#alur', label: 'Alur' },
  { href: '/#kemampuan', label: 'Kemampuan' },
  { href: '/#pengguna', label: 'Pengguna' },
  { href: '/faq', label: 'FAQ' },
]

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  return (
    <div className="simtas-dark flex min-h-screen flex-col">
      <NavbarSection navLinks={NAV_LINKS} />

      <main className="flex-1">{children}</main>

      <FooterSection navLinks={NAV_LINKS} />
    </div>
  )
}
