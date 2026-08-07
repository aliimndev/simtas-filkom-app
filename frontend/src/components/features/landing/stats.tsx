const STATS = [
  { value: '5', label: 'Peran pengguna' },
  { value: '6', label: 'Tahap Tugas Akhir' },
  { value: '1', label: 'Fakultas Ilmu Komputer' },
]

export function StatsSection() {
  return (
    <section className="border-b border-st-stroke">
      <div className="landing-container grid grid-cols-3 divide-x divide-st-stroke">
        {STATS.map((s) => (
          <div key={s.label} className="px-3 py-10 text-center md:py-12">
            <p className="font-display text-4xl text-st-text md:text-6xl">{s.value}</p>
            <p className="mt-2 text-[0.7rem] uppercase tracking-[0.2em] text-st-muted md:text-xs">
              {s.label}
            </p>
          </div>
        ))}
      </div>
    </section>
  )
}
