import { GraduationCap, LayoutDashboard, ShieldCheck, Users, type LucideIcon } from 'lucide-react'
import { Reveal } from './reveal'

interface RoleCard {
  icon: LucideIcon
  role: string
  desc: string
}

const ROLE_CARDS: RoleCard[] = [
  { icon: GraduationCap, role: 'Mahasiswa', desc: 'Mengajukan judul, mengunggah dokumen, dan memantau bimbingan.' },
  { icon: Users, role: 'Dosen Pembimbing & Penguji', desc: 'Memvalidasi bimbingan, mereview dokumen, dan menilai.' },
  { icon: ShieldCheck, role: 'Kaprodi', desc: 'Menyetujui judul, memantau progres, dan mengelola penjadwalan.' },
  { icon: LayoutDashboard, role: 'Administrator Fakultas', desc: 'Mengelola pengguna, konfigurasi sistem, dan data master.' },
]

export function RolesSection() {
  return (
    <section className="border-b border-st-stroke">
      <div className="landing-container py-20 md:py-28">
        <Reveal>
          <span className="landing-eyebrow">
            <span className="h-px w-8 bg-st-stroke" /> Pengguna
          </span>
        </Reveal>
        <Reveal delay={80}>
          <h2 className="landing-display mt-5 max-w-2xl text-3xl md:text-5xl">
            Dibuat untuk setiap <span className="accent-text italic">peran</span>.
          </h2>
        </Reveal>

        <div className="mt-14 grid gap-4 sm:grid-cols-2">
          {ROLE_CARDS.map((p, i) => {
            const Icon = p.icon
            return (
              <Reveal key={p.role} delay={i * 60} className="h-full">
                <div className="st-card flex h-full items-start gap-5 rounded-2xl p-6 md:p-7">
                  <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl accent-gradient text-st-bg">
                    <Icon className="h-5 w-5" />
                  </span>
                  <div>
                    <h3 className="text-base font-medium text-st-text">{p.role}</h3>
                    <p className="mt-1.5 text-sm leading-relaxed text-st-muted">{p.desc}</p>
                  </div>
                </div>
              </Reveal>
            )
          })}
        </div>
      </div>
    </section>
  )
}
