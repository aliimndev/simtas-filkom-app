'use client'

import Link from 'next/link'

type NavLink = { href: string; label: string }

export function FooterSection({ navLinks }: { navLinks: NavLink[] }) {
  return (
    <footer className="border-t border-st-stroke">
      <div className="landing-container grid gap-12 py-16 md:grid-cols-4">
        <div className="md:col-span-2">
          <div className="flex items-center gap-2.5">
            <span className="accent-ring flex h-8 w-8 items-center justify-center rounded-full">
              <span className="flex h-[calc(100%-3px)] w-[calc(100%-3px)] items-center justify-center rounded-full bg-st-bg">
                <span className="font-display text-[13px] italic text-st-text">
                  sf
                </span>
              </span>
            </span>
            <span className="text-sm font-medium tracking-tight text-st-text">
              SIMTAS <span className="text-st-muted">FILKOM</span>
            </span>
          </div>
          <p className="mt-5 max-w-sm text-sm leading-relaxed text-st-muted">
            Sistem Manajemen Tugas Akhir Skripsi Fakultas Ilmu Komputer
            Universitas Djuanda — mengelola seluruh proses Tugas Akhir dari
            pengajuan judul, bimbingan, seminar, hingga sidang dan arsip.
          </p>
          <p className="mt-4 text-sm">
            <a href="mailto:simtas@filkom.unida.ac.id" className="text-st-muted transition hover:text-st-text">
              simtas@filkom.unida.ac.id
            </a>
          </p>
        </div>

        <div>
          <h4 className="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-st-muted">
            Navigasi
          </h4>
          <ul className="mt-5 space-y-3 text-sm">
            {navLinks.map((l) => (
              <li key={l.href}>
                <Link href={l.href} className="text-st-muted transition hover:text-st-text">
                  {l.label}
                </Link>
              </li>
            ))}
          </ul>
        </div>

        <div>
          <h4 className="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-st-muted">
            Akses Sistem
          </h4>
          <ul className="mt-5 space-y-3 text-sm">
            <li>
              <Link href="/login" className="text-st-muted transition hover:text-st-text">
                Masuk
              </Link>
            </li>
            <li>
              <Link href="/forgot-password" className="text-st-muted transition hover:text-st-text">
                Lupa Password
              </Link>
            </li>
          </ul>
        </div>
      </div>

      <div className="border-t border-st-stroke">
        <div className="landing-container flex flex-col items-center justify-between gap-2 py-6 text-xs text-st-muted sm:flex-row">
          <p>© {new Date().getFullYear()} SIMTAS FILKOM — Universitas Djuanda.</p>
          <p>Fakultas Ilmu Komputer · Universitas Djuanda</p>
        </div>
      </div>
    </footer>
  )
}
