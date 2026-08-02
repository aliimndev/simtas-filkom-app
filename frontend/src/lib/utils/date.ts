const MONTHS_ID = [
  'Jan', 'Feb', 'Mar', 'Apr', 'Mei', 'Jun',
  'Jul', 'Agu', 'Sep', 'Okt', 'Nov', 'Des',
]

/**
 * Ubah string tanggal menjadi Date dengan interpretasi WAKTU LOKAL.
 * `new Date("2026-08-12")` di JS di-parse sebagai UTC tengah malam, sehingga
 * di WIB (UTC+7) tampil sehari lebih awal. Tanggal tanpa waktu dianggap lokal.
 */
function toLocalDate(value: string | Date): Date {
  if (value instanceof Date) return value
  // YYYY-MM-DD (tanpa waktu) → parse sebagai lokal
  if (/^\d{4}-\d{2}-\d{2}$/.test(value)) {
    const [y, m, d] = value.split('-').map(Number)
    return new Date(y, m - 1, d)
  }
  return new Date(value)
}

/** Format tanggal ke "12 Agu 2026" (Indonesia). */
export function formatDate(value?: string | Date | null): string {
  if (!value) return '—'
  const d = toLocalDate(value)
  if (Number.isNaN(d.getTime())) return '—'
  return `${d.getDate()} ${MONTHS_ID[d.getMonth()]} ${d.getFullYear()}`
}

/** Format tanggal-waktu ke "12 Agu 2026, 14:30". */
export function formatDateTime(value?: string | Date | null): string {
  if (!value) return '—'
  const d = toLocalDate(value)
  if (Number.isNaN(d.getTime())) return '—'
  const hh = String(d.getHours()).padStart(2, '0')
  const mm = String(d.getMinutes()).padStart(2, '0')
  return `${formatDate(d)}, ${hh}:${mm}`
}

/** Relative "x hari lagi" untuk countdown jadwal. */
export function daysUntil(value?: string | Date | null): string {
  if (!value) return '—'
  const target = toLocalDate(value).getTime()
  if (Number.isNaN(target)) return '—'
  const diff = Math.ceil((target - Date.now()) / (1000 * 60 * 60 * 24))
  if (diff < 0) return 'Lewat'
  if (diff === 0) return 'Hari ini'
  if (diff === 1) return 'Besok'
  return `${diff} hari lagi`
}
