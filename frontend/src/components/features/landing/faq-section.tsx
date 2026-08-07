import { Reveal } from './reveal'
import { FaqAccordion, type FaqItem } from './faq-accordion'

const FAQ_ITEMS: FaqItem[] = [
  { q: 'Bagaimana cara login ke sistem?', a: 'Buka halaman Masuk, lalu gunakan alamat email dan password yang diberikan oleh administrator fakultas.' },
  { q: 'Bagaimana jika saya lupa password?', a: 'Gunakan menu Lupa Password, masukkan email Anda, lalu ikuti tautan reset yang dikirim melalui email.' },
  { q: 'Bagaimana cara mengajukan judul?', a: 'Masuk sebagai mahasiswa, buka menu Tugas Akhir Skripsi, lalu isi formulir pengajuan judul dan kirimkan untuk direview.' },
  { q: 'Apakah saya bisa melihat jadwal seminar dan sidang?', a: 'Ya. Jadwal tampil pada dashboard dan menu Jadwal sesuai peran masing-masing.' },
]

export function FaqSection() {
  return (
    <section className="landing-surface border-b border-st-stroke">
      <div className="landing-container py-20 md:py-28">
        <Reveal>
          <div className="max-w-2xl">
            <span className="landing-eyebrow">FAQ</span>
            <h2 className="landing-display mt-5 text-3xl md:text-5xl">Pertanyaan yang sering muncul.</h2>
          </div>
        </Reveal>
        <div className="mt-10 max-w-2xl">
          <FaqAccordion items={FAQ_ITEMS} variant="list" />
        </div>
      </div>
    </section>
  )
}
