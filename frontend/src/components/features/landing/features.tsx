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

export function FeaturesSection() {
  return (
    <section
      id="fitur"
      aria-labelledby="kemampuan-title"
      className="landing-surface relative scroll-mt-24 overflow-hidden"
    >
      {/* Soft background glow */}
      <div
        aria-hidden
        className="pointer-events-none absolute inset-x-0 top-0 h-80 bg-linear-to-b from-(--st-accent-from) via-transparent to-transparent opacity-[0.05]"
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
              className="landing-display mt-5 text-balance text-3xl md:text-5xl"
            >
              Semua proses dalam satu sistem.
            </h2>
          </Reveal>

          <Reveal delay={120}>
            <p className="mt-5 max-w-2xl text-sm leading-7 text-st-muted md:text-base">
              SIMTAS membantu setiap peran mengelola proses Tugas Akhir sesuai kebutuhan dan
              kewenangannya.
            </p>
          </Reveal>
        </div>

        <div className="mt-14 grid gap-4 md:grid-cols-6">
          {FEATURES.map((feature, index) => {
            const Icon = feature.icon
            const number = String(index + 1).padStart(2, '0')
            const isPrimary = index === 0
            const isSecondary = index > 0 && index < 3
            const spanClass = isPrimary
              ? 'md:col-span-3 md:row-span-2'
              : isSecondary
                ? 'md:col-span-3'
                : 'md:col-span-2'

            return (
              <Reveal key={feature.title} delay={index * 50} className={`h-full ${spanClass}`}>
                <article className="group relative flex h-full flex-col overflow-hidden rounded-2xl border border-st-stroke bg-st-surface/80 p-6 backdrop-blur-sm transition hover:bg-st-surface-hi md:p-7">
                  <div
                    aria-hidden
                    className="pointer-events-none absolute -right-12 -bottom-12 h-36 w-36 rounded-full bg-(--st-accent-from)/10 blur-2xl transition-opacity duration-500 group-hover:opacity-100"
                  />
                  <div className="relative flex items-center justify-between gap-4">
                    <span className="transition-transform duration-300 group-hover:-translate-y-0.5">
                      <Icon
                        className={`text-(--st-accent-from) transition-colors duration-300 group-hover:text-(--st-accent-to) ${
                          isPrimary ? 'h-7 w-7' : 'h-6 w-6'
                        }`}
                        aria-hidden
                      />
                    </span>

                    <span
                      aria-hidden
                      className="select-none text-4xl font-semibold tracking-tight text-st-text opacity-[0.08] transition-opacity duration-300 group-hover:opacity-20"
                    >
                      {number}
                    </span>
                  </div>

                  <h3
                    className={`mt-5 max-w-xs font-semibold leading-snug text-st-text transition-colors duration-300 group-hover:text-(--st-accent-from) ${
                      isPrimary ? 'text-xl md:text-2xl' : 'text-lg'
                    }`}
                  >
                    {feature.title}
                  </h3>

                  <p
                    className={`mt-3 max-w-sm text-st-muted ${
                      isPrimary ? 'text-sm leading-7 md:text-base' : 'text-sm leading-7'
                    }`}
                  >
                    {feature.desc}
                  </p>
                </article>
              </Reveal>
            )
          })}
        </div>

        <div className="mt-16 border-t border-st-stroke pt-12">
          <Reveal>
            <span className="landing-eyebrow">PENGGUNA</span>
          </Reveal>
          <Reveal delay={80}>
            <h3 className="landing-display mt-5 max-w-2xl text-2xl md:text-4xl">
              Dibuat untuk setiap peran.
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
