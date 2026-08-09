import type { Metadata } from 'next'
import Script from 'next/script'
import './globals.css'
import { Providers } from '@/providers'

export const metadata: Metadata = {
  title: 'SIMTAS FILKOM',
  description: 'Sistem Manajemen Tugas Akhir Skripsi FILKOM Unida',
}

/*
Apply the persisted theme before React hydrates so returning dark-mode users
never see a light flash. A React-rendered inline <script> would trip React 19's
"Encountered a script tag" dev warning, so this is injected via next/script
(beforeInteractive) instead. color-scheme is handled by the `.dark` class and
the CSS rules in globals.css.
*/
const THEME_INIT_SCRIPT = `(() => {
  try {
    if (localStorage.getItem('theme') === 'dark') {
      document.documentElement.classList.add('dark');
    }
  } catch {}
})()`

export default function RootLayout({ children }: { children: React.ReactNode }) {
  return (
    <html lang="id" suppressHydrationWarning>
      <body>
        <Script id="theme-init" strategy="beforeInteractive">
          {THEME_INIT_SCRIPT}
        </Script>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
