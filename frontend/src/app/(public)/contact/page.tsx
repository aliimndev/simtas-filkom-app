import { MapPin, Mail, Phone, Clock } from 'lucide-react'
import { Card, CardContent } from '@/components/ui/card'

const INFO = [
  {
    icon: MapPin,
    title: 'Alamat',
    lines: ['Fakultas Ilmu Komputer, Universitas Djuanda', 'Kampus Ciawi, Bogor, Jawa Barat, Indonesia'],
  },
  {
    icon: Mail,
    title: 'Email',
    lines: ['filkom@unida.ac.id', 'simtas@filkom.unida.ac.id'],
  },
  {
    icon: Phone,
    title: 'Telepon',
    lines: ['(0251) 1234 5678', '+62 812 3456 7890'],
  },
  {
    icon: Clock,
    title: 'Jam Operasional',
    lines: ['Senin – Jumat', '08.00 – 16.00 WIB'],
  },
]

export const metadata = { title: 'Kontak — SIMTAS FILKOM' }

export default function ContactPage() {
  return (
    <div>
      <section className="border-b border-border">
        <div className="landing-container max-w-3xl py-16">
          <span className="landing-eyebrow">
            <Mail className="h-4 w-4" /> Kontak
          </span>
          <h1 className="landing-display mt-4 text-4xl md:text-5xl">Hubungi kami</h1>
          <p className="mt-5 text-lg text-muted-foreground">
            Ada pertanyaan atau kendala teknis? Silakan hubungi kami melalui kanal berikut.
          </p>
        </div>
      </section>

      <section className="landing-surface border-b border-border">
        <div className="landing-container py-16">
          <div className="grid gap-5 sm:grid-cols-2">
            {INFO.map((c) => (
              <Card key={c.title}>
                <CardContent className="flex gap-4 p-6">
                  <span className="flex h-11 w-11 shrink-0 items-center justify-center rounded-lg bg-primary-50 text-primary">
                    <c.icon className="h-5 w-5" />
                  </span>
                  <div>
                    <h3 className="font-semibold text-foreground">{c.title}</h3>
                    {c.lines.map((l) => (
                      <p key={l} className="mt-1 text-sm text-muted-foreground">
                        {l}
                      </p>
                    ))}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="mt-6 rounded-xl border border-dashed border-border bg-muted/40 p-8 text-center text-sm text-muted-foreground">
            <p>
              Untuk layanan administratif Tugas Akhir, kunjungi sekretariat fakultas pada jam operasional di atas,
              atau sampaikan kendala teknis aplikasi melalui email{' '}
              <a href="mailto:simtas@filkom.unida.ac.id" className="font-medium text-primary hover:underline">
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
