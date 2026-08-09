import { ArrowRight } from 'lucide-react'
import { Reveal } from '@/components/features/landing/reveal'
import { FaqAccordion, type FaqItem } from '@/components/features/landing/faq-accordion'

const FAQ_ITEMS: FaqItem[] = [
  { q: 'Bagaimana cara login ke sistem?', a: 'Login hanya menggunakan email dan password yang diberikan oleh administrator fakultas. Buka halaman Masuk, lalu masukkan kredensial Anda.' },
  { q: 'Bagaimana cara reset password?', a: 'Klik “Lupa Password” pada halaman login, masukkan email Anda, lalu ikuti tautan reset yang dikirim melalui email untuk mengatur password baru.' },
  { q: 'Bagaimana cara mengajukan judul?', a: 'Masuk sebagai mahasiswa, buka menu Tugas Akhir Skripsi, klik Ajukan Judul, lengkapi formulir, dan kirimkan untuk direview oleh kaprodi.' },
  { q: 'Bagaimana mengetahui dosen pembimbing?', a: 'Setelah judul disetujui, sistem akan menugaskan dosen pembimbing. Daftar pembimbing dapat dilihat pada detail Tugas Akhir di dashboard.' },
  { q: 'Bagaimana melihat jadwal seminar?', a: 'Jadwal seminar dan sidang tampil pada dashboard dan menu Jadwal sesuai peran Anda masing-masing.' },
  { q: 'Bagaimana cara mengunggah revisi?', a: 'Buka menu Dokumen pada Tugas Akhir, unggah versi revisi dengan jenis dokumen yang sesuai, lalu tunggu review dari dosen pembimbing.' },
  { q: 'Bagaimana jika proposal ditolak?', a: 'Jika proposal ditolak, Anda akan melihat catatan revisi dari kaprodi. Perbaiki sesuai catatan lalu ajukan ulang.' },
]

export const metadata = { title: 'FAQ — SIMTAS FILKOM' }

export default function FaqPage() {
  return (
    <div>
      <section className="border-b border-st-stroke">
        <div className="landing-container max-w-3xl pb-16 pt-36 md:pt-44">
          <Reveal><span className="landing-eyebrow">FAQ</span></Reveal>
          <Reveal delay={80}><h1 className="landing-display mt-5 text-4xl md:text-5xl">Pertanyaan yang sering diajukan</h1></Reveal>
          <Reveal delay={160}>
            <p className="mt-5 text-base text-st-muted md:text-lg">
              Temukan jawaban atas pertanyaan umum seputar penggunaan sistem.
            </p>
          </Reveal>
        </div>
      </section>

      <section className="landing-surface border-b border-st-stroke">
        <div className="landing-container max-w-3xl py-16">
          <FaqAccordion items={FAQ_ITEMS} variant="cards" />

          <div className="mt-12 rounded-2xl border border-st-stroke bg-st-surface p-6 text-center">
            <p className="font-medium text-st-text">Tidak menemukan jawaban yang Anda cari?</p>
            <a href="mailto:simtas@filkom.unida.ac.id" className="accent-ring mt-3 inline-flex items-center gap-2 rounded-full bg-st-text px-5 py-2.5 text-sm font-medium text-st-bg transition hover:opacity-90">
              Hubungi Kami <ArrowRight className="h-4 w-4" />
            </a>
          </div>
        </div>
      </section>
    </div>
  )
}
