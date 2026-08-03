'use client'

import { useState } from 'react'
import Link from 'next/link'
import { ChevronDown, ArrowRight } from 'lucide-react'
import { Reveal } from '@/components/features/landing/reveal'

const FAQ_ITEMS = [
  { q: 'Bagaimana cara login ke sistem?', a: 'Login hanya menggunakan email dan password yang diberikan oleh administrator fakultas. Buka halaman Masuk, lalu masukkan kredensial Anda.' },
  { q: 'Bagaimana cara reset password?', a: 'Klik “Lupa Password” pada halaman login, masukkan email Anda, lalu ikuti tautan reset yang dikirim melalui email untuk mengatur password baru.' },
  { q: 'Bagaimana cara mengajukan judul?', a: 'Masuk sebagai mahasiswa, buka menu Tugas Akhir Skripsi, klik Ajukan Judul, lengkapi formulir, dan kirimkan untuk direview oleh kaprodi.' },
  { q: 'Bagaimana mengetahui dosen pembimbing?', a: 'Setelah judul disetujui, sistem akan menugaskan dosen pembimbing. Daftar pembimbing dapat dilihat pada detail Tugas Akhir di dashboard.' },
  { q: 'Bagaimana melihat jadwal seminar?', a: 'Jadwal seminar dan sidang tampil pada dashboard dan menu Jadwal sesuai peran Anda masing-masing.' },
  { q: 'Bagaimana cara mengunggah revisi?', a: 'Buka menu Dokumen pada Tugas Akhir, unggah versi revisi dengan jenis dokumen yang sesuai, lalu tunggu review dari dosen pembimbing.' },
  { q: 'Bagaimana jika proposal ditolak?', a: 'Jika proposal ditolak, Anda akan melihat catatan revisi dari kaprodi. Perbaiki sesuai catatan lalu ajukan ulang.' },
]

export default function FaqPage() {
  const [open, setOpen] = useState<number | null>(0)

  return (
    <div>
      <section className="border-b border-[var(--st-stroke)]">
        <div className="landing-container max-w-3xl pb-16 pt-36 md:pt-44">
          <Reveal><span className="landing-eyebrow">FAQ</span></Reveal>
          <Reveal delay={80}><h1 className="landing-display mt-5 text-4xl md:text-5xl">Pertanyaan yang sering diajukan</h1></Reveal>
          <Reveal delay={160}>
            <p className="mt-5 text-base text-[var(--st-muted)] md:text-lg">
              Temukan jawaban atas pertanyaan umum seputar penggunaan sistem.
            </p>
          </Reveal>
        </div>
      </section>

      <section className="landing-surface border-b border-[var(--st-stroke)]">
        <div className="landing-container max-w-3xl py-16">
          <div className="space-y-3">
            {FAQ_ITEMS.map((item, i) => (
              <div key={item.q} className="overflow-hidden rounded-2xl border border-[var(--st-stroke)] bg-[var(--st-surface)]">
                <button
                  type="button"
                  onClick={() => setOpen(open === i ? null : i)}
                  className="flex w-full items-center justify-between gap-4 px-5 py-4 text-left"
                >
                  <span className="font-medium text-[var(--st-text)]">{item.q}</span>
                  <ChevronDown className={`h-4 w-4 shrink-0 text-[var(--st-muted)] transition-transform ${open === i ? 'rotate-180' : ''}`} />
                </button>
                {open === i && (
                  <div className="border-t border-[var(--st-stroke)] px-5 py-4 text-sm leading-relaxed text-[var(--st-muted)]">
                    {item.a}
                  </div>
                )}
              </div>
            ))}
          </div>

          <div className="mt-12 rounded-2xl border border-[var(--st-stroke)] bg-[var(--st-surface)] p-6 text-center">
            <p className="font-medium text-[var(--st-text)]">Tidak menemukan jawaban yang Anda cari?</p>
            <Link href="/contact" className="accent-ring mt-3 inline-flex items-center gap-2 rounded-full bg-[var(--st-text)] px-5 py-2.5 text-sm font-medium text-[var(--st-bg)] transition hover:opacity-90">
              Hubungi Kami <ArrowRight className="h-4 w-4" />
            </Link>
          </div>
        </div>
      </section>
    </div>
  )
}
