'use client'

import { useState } from 'react'
import Link from 'next/link'
import { HelpCircle, ChevronDown, ArrowRight } from 'lucide-react'

const FAQ_ITEMS = [
  {
    q: 'Bagaimana cara login ke sistem?',
    a: 'Login hanya menggunakan email dan password yang diberikan oleh administrator fakultas. Buka halaman Masuk, lalu masukkan kredensial Anda.',
  },
  {
    q: 'Bagaimana cara reset password?',
    a: 'Klik "Lupa Password" pada halaman login, masukkan email Anda, lalu ikuti tautan reset yang dikirim melalui email untuk mengatur password baru.',
  },
  {
    q: 'Bagaimana cara mengajukan judul?',
    a: 'Masuk sebagai mahasiswa, buka menu Skripsi Saya, klik Ajukan Judul, lengkapi formulir, dan kirimkan untuk direview oleh kaprodi.',
  },
  {
    q: 'Bagaimana mengetahui dosen pembimbing?',
    a: 'Setelah judul disetujui, sistem akan menugaskan dosen pembimbing. Daftar pembimbing dapat dilihat pada detail skripsi di dashboard.',
  },
  {
    q: 'Bagaimana melihat jadwal seminar?',
    a: 'Jadwal seminar dan sidang tampil pada dashboard dan menu Jadwal sesuai peran Anda masing-masing.',
  },
  {
    q: 'Bagaimana cara mengunggah revisi?',
    a: 'Buka menu Dokumen pada skripsi, unggah versi revisi dengan jenis dokumen yang sesuai, lalu tunggu review dari dosen pembimbing.',
  },
  {
    q: 'Bagaimana jika proposal ditolak?',
    a: 'Jika proposal ditolak, Anda akan melihat catatan revisi dari kaprodi. Perbaiki sesuai catatan lalu ajukan ulang.',
  },
]

export default function FaqPage() {
  const [open, setOpen] = useState<number | null>(0)

  return (
    <div>
      <section className="border-b border-border">
        <div className="landing-container max-w-3xl py-16">
          <span className="landing-eyebrow">
            <HelpCircle className="h-4 w-4" /> FAQ
          </span>
          <h1 className="landing-display mt-4 text-4xl md:text-5xl">Pertanyaan yang sering diajukan</h1>
          <p className="mt-5 text-lg text-muted-foreground">
            Temukan jawaban atas pertanyaan umum seputar penggunaan sistem.
          </p>
        </div>
      </section>

      <section className="landing-surface border-b border-border">
        <div className="landing-container max-w-3xl py-16">
          <div className="space-y-3">
            {FAQ_ITEMS.map((item, i) => (
              <div key={item.q} className="overflow-hidden rounded-xl border border-border bg-background">
                <button
                  type="button"
                  onClick={() => setOpen(open === i ? null : i)}
                  className="flex w-full items-center justify-between gap-4 px-5 py-4 text-left"
                >
                  <span className="font-medium text-foreground">{item.q}</span>
                  <ChevronDown
                    className={`h-5 w-5 shrink-0 text-muted-foreground transition-transform ${open === i ? 'rotate-180' : ''}`}
                  />
                </button>
                {open === i && (
                  <div className="border-t border-border px-5 py-4 text-sm leading-relaxed text-muted-foreground">
                    {item.a}
                  </div>
                )}
              </div>
            ))}
          </div>

          <div className="mt-12 rounded-xl border border-border bg-background p-6 text-center">
            <p className="font-semibold text-foreground">Tidak menemukan jawaban yang Anda cari?</p>
            <Link
              href="/contact"
              className="mt-3 inline-flex items-center gap-2 rounded-lg bg-primary px-5 py-2.5 text-sm font-medium text-primary-foreground transition hover:bg-primary-700"
            >
              Hubungi Kami <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>
    </div>
  )
}
