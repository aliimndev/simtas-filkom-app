import type { Metadata } from 'next'
import Script from 'next/script'
import { Instrument_Serif, Inter } from 'next/font/google'
import './globals.css'
import { Providers } from '@/providers'

/*
Self-host the design typefaces (Inter for body, Instrument Serif for display)
via next/font instead of the external Google Fonts <link> in globals.css. This
bundles the font files locally so the site renders the intended typeface even
when the Google Fonts CDN is unreachable (the cause of "default font" locally).
The exposed CSS vars keep every existing --font-body / --font-display usage.
*/
const fontBody = Inter({
  subsets: ['latin'],
  variable: '--font-body',
  display: 'swap',
})

const fontDisplay = Instrument_Serif({
  weight: '400',
  style: ['normal', 'italic'],
  subsets: ['latin'],
  variable: '--font-display',
  display: 'swap',
})

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
    <html lang="id" suppressHydrationWarning className={`${fontBody.variable} ${fontDisplay.variable}`}>
      <body>
        <Script id="theme-init" strategy="beforeInteractive">
          {THEME_INIT_SCRIPT}
        </Script>
        <Providers>{children}</Providers>
      </body>
    </html>
  )
}
