import { MapPin, Mail, Phone, Clock } from 'lucide-react'
import { Reveal } from '@/components/features/landing/reveal'

const INFO = [
  { icon: MapPin, title: 'Alamat', lines: ['Fakultas Ilmu Komputer, Universitas Djuanda', 'Kampus Ciawi, Bogor, Jawa Barat, Indonesia'] },
  { icon: Mail, title: 'Email', lines: ['filkom@unida.ac.id', 'simtas@filkom.unida.ac.id'] },
  { icon: Phone, title: 'Telepon', lines: ['(0251) 1234 5678', '+62 812 3456 7890'] },
  { icon: Clock, title: 'Jam Operasional', lines: ['Senin – Jumat', '08.00 – 16.00 WIB'] },
]

export const metadata = { title: 'Kontak — SIMTAS FILKOM' }

export default function ContactPage() {
  return (
    <div>
      <section className="border-b border-st-stroke">
        <div className="landing-container max-w-3xl pb-16 pt-36 md:pt-44">
          <Reveal><span className="landing-eyebrow">Kontak</span></Reveal>
          <Reveal delay={80}><h1 className="landing-display mt-5 text-4xl md:text-5xl">Hubungi kami</h1></Reveal>
          <Reveal delay={160}>
            <p className="mt-5 text-base text-st-muted md:text-lg">
              Ada pertanyaan atau kendala teknis? Silakan hubungi kami melalui kanal berikut.
            </p>
          </Reveal>
        </div>
      </section>

      <section className="landing-surface border-b border-st-stroke">
        <div className="landing-container py-16">
          <div className="grid gap-4 sm:grid-cols-2">
            {INFO.map((c, i) => {
              const Icon = c.icon
              return (
                <Reveal key={c.title} delay={i * 60} className="h-full">
                  <div className="st-card flex h-full gap-4 rounded-2xl p-6">
                    <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl accent-gradient text-st-bg">
                      <Icon className="h-5 w-5" />
                    </span>
                    <div>
                      <h3 className="font-medium text-st-text">{c.title}</h3>
                      {c.lines.map((l) => (
                        <p key={l} className="mt-1 text-sm text-st-muted">{l}</p>
                      ))}
                    </div>
                  </div>
                </Reveal>
              )
            })}
          </div>

          <div className="mt-6 rounded-2xl border border-dashed border-st-stroke p-8 text-center text-sm text-st-muted">
            <p>
              Untuk layanan administratif Tugas Akhir, kunjungi sekretariat fakultas pada jam operasional di atas,
              atau sampaikan kendala teknis aplikasi melalui email{' '}
              <a href="mailto:simtas@filkom.unida.ac.id" className="font-medium text-(--st-accent-from) hover:underline">
                simtas@filkom.unida.ac.id
              </a>
              .
            </p>
          </div>
        </div>
      </section>
    </div>
  )
}
