'use client'

import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { thesisApi } from '@/lib/api/thesis-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { Select } from '@/components/ui/select'
import type { Thesis } from '@/types/thesis'

/**
 * Hook untuk halaman yang bekerja pada satu skripsi (dokumen, bimbingan).
 * - Mahasiswa: otomatis memakai skripsi miliknya.
 * - Dosen/Kaprodi/Admin: memuat daftar skripsi (role-scoped dari backend) dan
 *   menampilkan <Select> agar bisa memilih mahasiswa bimbingan mana yang dilihat.
 */
export function useThesisPicker() {
  const user = useAuthStore((s) => s.user)
  const isStudent = user?.role === 'mahasiswa'
  const [selectedId, setSelectedId] = useState<string | null>(null)

  const theses = useQuery({
    queryKey: ['theses', 'list-for-picker'],
    queryFn: () => thesisApi.list({ per_page: 100 }),
    enabled: Boolean(user),
  })

  const all = theses.data?.data ?? []

  // Mahasiswa hanya punya 1 skripsi (role-scoped); staf bisa punya banyak.
  const thesis: Thesis | undefined = isStudent
    ? all[0]
    : all.find((t) => t.id === selectedId) ?? all[0]

  const picker =
    !isStudent && all.length > 1 ? (
      <Select
        value={thesis?.id ?? ''}
        onChange={(e) => setSelectedId(e.target.value)}
        className="max-w-xs"
        aria-label="Pilih skripsi mahasiswa"
      >
        {all.map((t) => (
          <option key={t.id} value={t.id}>
            {(t.student?.full_name ?? '-') + ' - ' + t.title.slice(0, 40)}
            {t.title.length > 40 ? '...' : ''}
          </option>
        ))}
      </Select>
    ) : null

  return { thesis, picker, isLoading: theses.isLoading }
}
