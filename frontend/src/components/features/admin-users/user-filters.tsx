import { Search } from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { ROLE_OPTIONS } from '@/constants/roles'

/** Pencarian nama/email/NIM + filter peran pengguna. */
export function UserFilters({
  q,
  role,
  onQChange,
  onRoleChange,
}: {
  q: string
  role: string
  onQChange: (q: string) => void
  onRoleChange: (role: string) => void
}) {
  return (
    <div className="flex flex-wrap items-center gap-3">
      <div className="relative">
        <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
        <Input
          placeholder="Cari nama / email / NIM…"
          value={q}
          onChange={(e) => onQChange(e.target.value)}
          className="max-w-xs pl-9"
        />
      </div>
      <Select value={role} onChange={(e) => onRoleChange(e.target.value)} className="w-44">
        <option value="">Semua peran</option>
        {ROLE_OPTIONS.map((r) => (
          <option key={r.value} value={r.value}>{r.label}</option>
        ))}
      </Select>
    </div>
  )
}
