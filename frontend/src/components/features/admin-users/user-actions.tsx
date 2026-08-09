import { KeyRound, Power, PowerOff } from 'lucide-react'
import { Button } from '@/components/ui/button'
import type { User } from '@/types/auth'

/** Aksi baris pengguna (aktif/nonaktif + reset password) — dipakai tabel & kartu mobile. */
export function UserActions({
  user,
  onToggleActive,
  onResetPassword,
  layout = 'col',
}: {
  user: User
  onToggleActive: (user: User) => void
  onResetPassword: (user: User) => void
  layout?: 'row' | 'col'
}) {
  return (
    <div className={layout === 'row' ? 'flex justify-end gap-1' : 'flex flex-col gap-1.5'}>
      <Button
        size="sm"
        variant="ghost"
        title={user.is_active === false ? 'Aktifkan' : 'Nonaktifkan'}
        aria-label={user.is_active === false ? 'Aktifkan pengguna' : 'Nonaktifkan pengguna'}
        onClick={() => onToggleActive(user)}
      >
        {user.is_active === false ? <Power className="h-4 w-4 text-success" /> : <PowerOff className="h-4 w-4 text-danger-700" />}
      </Button>
      <Button
        size="sm"
        variant="ghost"
        title="Reset password"
        aria-label="Reset password"
        onClick={() => onResetPassword(user)}
      >
        <KeyRound className="h-4 w-4" />
      </Button>
    </div>
  )
}
