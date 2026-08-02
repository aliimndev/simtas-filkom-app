import type { UserRole } from '@/types/auth'

export const ROLE_LABELS: Record<string, string> = {
  admin_fakultas: 'Admin Fakultas',
  kaprodi: 'Kaprodi',
  mahasiswa: 'Mahasiswa',
  dosen_pembimbing: 'Dosen Pembimbing',
  dosen_penguji: 'Dosen Penguji',
}

export const ROLE_OPTIONS: { value: string; label: string }[] = [
  { value: 'admin_fakultas', label: ROLE_LABELS.admin_fakultas },
  { value: 'kaprodi', label: ROLE_LABELS.kaprodi },
  { value: 'mahasiswa', label: ROLE_LABELS.mahasiswa },
  { value: 'dosen_pembimbing', label: ROLE_LABELS.dosen_pembimbing },
  { value: 'dosen_penguji', label: ROLE_LABELS.dosen_penguji },
]

export function roleLabel(role?: string | null): string {
  if (!role) return '—'
  return ROLE_LABELS[role] ?? role
}

export const STAFF_ROLES: UserRole[] = ['admin_fakultas', 'kaprodi']
export const SUPERVISOR_ROLES: UserRole[] = ['admin_fakultas', 'kaprodi', 'dosen_pembimbing']
