import {
  Archive,
  ArrowRight,
  Bell,
  CalendarCheck,
  FileText,
  LayoutDashboard,
  MessagesSquare,
  Search,
} from 'lucide-react'
import { ContainerScroll } from './container-scroll-animation'
import { LandingButton } from './landing-button'

const PREVIEW_STATS = [
  { label: 'Judul Diproses', value: '128' },
  { label: 'Bimbingan Aktif', value: '46' },
  { label: 'Seminar Terjadwal', value: '12' },
  { label: 'Arsip Final', value: '312' },
]

const PREVIEW_ROWS = [
  { title: 'Analisis Sentimen Kebijakan...', stage: 'Bimbingan', tint: 'text-(--st-accent-from)' },
  { title: 'Sistem Rekomendasi Mata...', stage: 'Seminar', tint: 'text-(--st-accent-to)' },
  { title: 'Deteksi Dini Plagiarisme...', stage: 'Sidang', tint: 'text-st-muted' },
]

function DashboardPreview() {
  return (
    // The preview is a decorative product mockup: it contains fake stats,
    // rows, and chrome that would only confuse a screen-reader user. Hide it
    // from assistive tech; the real content is the headline + CTA below.
    <div
      aria-hidden="true"
      className="grid h-full grid-cols-1 bg-st-bg text-left lg:grid-cols-[13.5rem_1fr]"
    >
      {/* Sidebar */}
      <aside className="hidden flex-col gap-1 border-r border-st-stroke p-5 lg:flex">
        <div className="mb-4 flex items-center gap-2.5 pl-1">
          <span className="accent-ring flex h-8 w-8 items-center justify-center rounded-full bg-st-surface">
            <span className="font-display text-[13px] italic text-st-text">sf</span>
          </span>
          <span className="text-sm font-medium tracking-tight text-st-text">
            SIMTAS <span className="text-st-muted">FILKOM</span>
          </span>
        </div>
        {[
          { icon: LayoutDashboard, label: 'Dashboard', active: true },
          { icon: FileText, label: 'Pengajuan' },
          { icon: MessagesSquare, label: 'Bimbingan' },
          { icon: CalendarCheck, label: 'Seminar' },
          { icon: Archive, label: 'Arsip' },
        ].map(({ icon: Icon, label, active }) => (
          <div
            key={label}
            className={`flex items-center gap-2.5 rounded-lg px-2.5 py-2 text-[13px] ${
              active
                ? 'bg-st-surface-hi font-medium text-(--st-accent-from)'
                : 'text-st-muted'
            }`}
          >
            <Icon className="h-4 w-4" aria-hidden />
            {label}
          </div>
        ))}
      </aside>

      {/* Main */}
      <div className="flex min-w-0 flex-col">
        {/* Topbar */}
        <div className="flex items-center justify-between gap-4 border-b border-st-stroke px-5 py-3">
          <div className="flex items-center gap-2 rounded-full border border-st-stroke bg-st-surface px-3 py-1.5 text-[13px] text-st-muted">
            <Search className="h-3.5 w-3.5" aria-hidden /> Cari judul skripsi…
          </div>
          <div className="flex items-center gap-3">
            <span className="relative flex h-8 w-8 items-center justify-center rounded-full border border-st-stroke bg-st-surface text-st-muted">
              <Bell className="h-3.5 w-3.5" aria-hidden />
              <span className="absolute right-1 top-1 h-1.5 w-1.5 rounded-full bg-(--st-accent-to)" />
            </span>
            <span className="flex h-8 w-8 items-center justify-center rounded-full bg-st-surface-hi font-display text-xs text-st-text">
              AR
            </span>
          </div>
        </div>

        {/* Stats */}
        <div className="grid grid-cols-2 gap-3 p-5 md:grid-cols-4">
          {PREVIEW_STATS.map((s) => (
            <div
              key={s.label}
              className="rounded-xl border border-st-stroke bg-st-surface p-4"
            >
              <p className="text-[11px] uppercase tracking-[0.15em] text-st-muted">{s.label}</p>
              <p className="mt-1.5 font-display text-2xl text-st-text tabular-nums">{s.value}</p>
            </div>
          ))}
        </div>

        {/* Table */}
        <div className="mx-5 mb-5 rounded-xl border border-st-stroke bg-st-surface">
          <div className="grid grid-cols-[1fr_auto] gap-4 border-b border-st-stroke px-4 py-2.5 font-mono text-[11px] uppercase tracking-[0.2em] text-st-muted">
            <span>Judul</span>
            <span>Tahap</span>
          </div>
          {PREVIEW_ROWS.map((row) => (
            <div
              key={row.title}
              className="grid grid-cols-[1fr_auto] items-center gap-4 border-b border-st-stroke px-4 py-3 last:border-0"
            >
              <span className="truncate text-[13px] text-st-text">{row.title}</span>
              <span className={`rounded-full border border-st-stroke px-2.5 py-1 text-[11px] ${row.tint}`}>
                {row.stage}
              </span>
            </div>
          ))}
        </div>
      </div>
    </div>
  )
}

export function HeroSection() {
  return (
    <section className="relative overflow-hidden">
      <div className="landing-container">
        <ContainerScroll
          titleComponent={
            <>
              <span className="landing-eyebrow">
                SIMTAS://FILKOM <span className="text-st-stroke">·</span> Universitas Djuanda
              </span>
<h1 className="mt-6 text-balance text-4xl font-semibold leading-[1.05] tracking-tight text-st-text md:text-6xl">
                Satu sistem untuk seluruh perjalanan{' '}
                <span className="accent-text italic md:text-7xl md:leading-[1.1]">
                  Tugas Akhir Skripsi.
                </span>
              </h1>
            </>
          }
        >
          <DashboardPreview />
        </ContainerScroll>

        <div className="pb-20 text-center md:pb-28">
          <LandingButton href="/login">
            Masuk ke Sistem <ArrowRight className="h-4 w-4" />
          </LandingButton>
        </div>
      </div>
    </section>
  )
}