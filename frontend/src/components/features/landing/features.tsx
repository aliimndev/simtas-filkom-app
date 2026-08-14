import {
  Archive,
  Bell,
  CalendarCheck,
  FileText,
  LayoutDashboard,
  MessagesSquare,
  ShieldCheck,
  UserCheck,
  GraduationCap,
  Users,
  type LucideIcon,
} from 'lucide-react'
import { Reveal } from './reveal'
import { NumberTicker } from './number-ticker'
import { ScrollStack, ScrollStackItem } from './scroll-stack'

interface Feature {
  icon: LucideIcon
  title: string
  desc: string
}

const FEATURES: Feature[] = [
  {
    icon: FileText,
    title: 'Pengajuan Judul',
    desc: 'Ajukan judul dan dapatkan review dari kaprodi secara digital.',
  },
  {
    icon: MessagesSquare,
    title: 'Bimbingan Online',
    desc: 'Catat log bimbingan dan menunggu validasi dosen pembimbing.',
  },
  {
    icon: CalendarCheck,
    title: 'Seminar & Sidang',
    desc: 'Penjadwalan, penguji, dan penilaian dalam satu alur.',
  },
  {
    icon: Archive,
    title: 'Arsip Digital',
    desc: 'Dokumen final terdokumentasi dan mudah diakses kembali.',
  },
  {
    icon: LayoutDashboard,
    title: 'Dashboard Real-time',
    desc: 'Statistik dan status per peran: mahasiswa, dosen, kaprodi, admin.',
  },
  {
    icon: Bell,
    title: 'Notifikasi Email',
    desc: 'Pemberitahuan penting terkirim ke pemangku kepentingan.',
  },
]

const ROLES: { icon: LucideIcon; title: string }[] = [
  { icon: GraduationCap, title: 'Mahasiswa' },
  { icon: UserCheck, title: 'Dosen Pembimbing' },
  { icon: Users, title: 'Dosen Penguji' },
  { icon: ShieldCheck, title: 'Kaprodi' },
  { icon: LayoutDashboard, title: 'Administrator Fakultas' },
]

const STATS: { value: number; suffix?: string; label: string }[] = [
  { value: 800, suffix: '+', label: 'Mahasiswa Terlayani' },
  { value: 1200, suffix: '+', label: 'Dokumen Tersimpan' },
  { value: 300, suffix: '+', label: 'Sidang Terproses' },
  { value: 48, label: 'Dosen Terbantu' },
]

export function FeaturesSection() {
  return (
    <section
      id="fitur"
      aria-labelledby="kemampuan-title"
      className="relative scroll-mt-24 overflow-hidden bg-st-surface"
    >
      {/* Feather the section band: the page bg color fades out over the top
          and bottom edge, so fitur reads as a distinct surface without a
          hard line where hero/pipeline meet it. */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-x-0 top-0 h-40 bg-linear-to-b from-st-bg to-transparent"
      />
      <div
        aria-hidden
        className="pointer-events-none absolute inset-x-0 bottom-0 h-40 bg-linear-to-t from-st-bg to-transparent"
      />
      {/* Soft accent wash gathering toward the content */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-x-0 top-40 h-80 bg-linear-to-b from-transparent to-(--st-accent-from) opacity-[0.05]"
      />

      <div className="landing-container relative py-20 md:py-28">
        {/* Header */}
        <div className="max-w-3xl">
          <Reveal>
            <span className="landing-eyebrow">
              <span className="h-1.5 w-1.5 rounded-full bg-(--st-accent-from)" />
              KEMAMPUAN
            </span>
          </Reveal>

          <Reveal delay={80}>
            <h2
              id="kemampuan-title"
              className="landing-heading mt-5 text-balance text-3xl md:text-5xl"
            >
              Semua proses dalam <span className="accent-text italic">satu sistem</span>.
            </h2>
          </Reveal>

          <Reveal delay={120}>
            <p className="mt-5 max-w-2xl text-sm leading-7 text-st-muted md:text-base">
              SIMTAS membantu setiap peran mengelola proses Tugas Akhir sesuai kebutuhan dan
              kewenangannya.
            </p>
          </Reveal>
        </div>

        <div className="mt-14">
          <ScrollStack
            itemDistance={40}
            itemStackDistance={14}
            itemScale={0.02}
            baseScale={0.92}
            stackPosition="18%"
            scaleEndPosition="12%"
          >
            {FEATURES.map((feature, index) => {
              const Icon = feature.icon
              const number = String(index + 1).padStart(2, '0')

              return (
                <ScrollStackItem key={feature.title}>
                  <article className="relative flex min-h-44 flex-col overflow-hidden rounded-2xl border border-st-stroke bg-st-surface p-6 md:min-h-52 md:p-8">
                    <div
                      aria-hidden
                      className="pointer-events-none absolute -right-16 -top-16 h-40 w-40 rounded-full bg-(--st-accent-from)/10 blur-2xl"
                    />
                    <div className="relative flex items-start justify-between gap-4">
                      <span className="accent-ring flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-st-surface-hi">
                        <Icon className="h-6 w-6 text-(--st-accent-from)" aria-hidden />
                      </span>
                      <span
                        aria-hidden
                        className="select-none font-display text-5xl font-semibold tracking-tight text-st-text opacity-[0.07] md:text-6xl"
                      >
                        {number}
                      </span>
                    </div>
                    <div className="relative mt-6 md:mt-8">
                      <h3 className="text-xl font-semibold tracking-tight text-st-text md:text-2xl">
                        {feature.title}
                      </h3>
                      <p className="mt-3 max-w-2xl text-sm leading-7 text-st-muted md:text-base">
                        {feature.desc}
                      </p>
                    </div>
                  </article>
                </ScrollStackItem>
              )
            })}
          </ScrollStack>
        </div>

        <div className="mt-16 grid grid-cols-2 gap-px overflow-hidden rounded-2xl border border-st-stroke bg-st-stroke sm:grid-cols-2 lg:grid-cols-4">
          {STATS.map((stat, index) => (
            <Reveal key={stat.label} delay={index * 70} className="h-full">
              <div className="flex h-full flex-col gap-1 bg-st-surface px-6 py-8 text-center transition-colors hover:bg-st-surface-hi md:py-10">
                <span className="font-display text-3xl text-(--st-accent-from) tabular-nums md:text-4xl">
                  <NumberTicker value={stat.value} />
                  {stat.suffix && <span className="text-(--st-accent-to)">{stat.suffix}</span>}
                </span>
                <span className="text-sm text-st-muted">{stat.label}</span>
              </div>
            </Reveal>
          ))}
        </div>

        <div className="mt-16 border-t border-st-stroke pt-12">
          <Reveal>
            <span className="landing-eyebrow">PENGGUNA</span>
          </Reveal>
          <Reveal delay={80}>
            <h3 className="landing-heading mt-5 max-w-2xl text-2xl md:text-4xl">
              Dibuat untuk <span className="accent-text italic">setiap peran</span>.
            </h3>
          </Reveal>
          <div className="mt-8 flex flex-wrap gap-3">
            {ROLES.map((role, index) => {
              const Icon = role.icon
              return (
                <Reveal key={role.title} delay={index * 45}>
                  <div className="inline-flex items-center gap-2 rounded-full border border-st-stroke bg-st-surface px-4 py-2 text-sm text-st-text transition hover:bg-st-surface-hi">
                    <Icon className="h-4 w-4 text-(--st-accent-from)" aria-hidden />
                    <span>{role.title}</span>
                  </div>
                </Reveal>
              )
            })}
          </div>
        </div>
      </div>
    </section>
  )
}
