import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { roleLabel } from '@/constants/roles'
import { UserActions } from './user-actions'
import type { User } from '@/types/auth'

/** Kartu pengguna (mobile). */
export function UserMobileCards({
  users,
  onToggleActive,
  onResetPassword,
}: {
  users: User[]
  onToggleActive: (user: User) => void
  onResetPassword: (user: User) => void
}) {
  return (
    <div className="space-y-3 md:hidden">
      {users.map((u) => (
        <Card key={u.id}>
          <CardContent className="p-4">
            <div className="flex items-start justify-between gap-3">
              <div className="min-w-0 flex-1">
                <p className="font-medium">{u.full_name}</p>
                <p className="mt-0.5 text-xs text-muted-foreground">{u.email}</p>
                <div className="mt-2 flex flex-wrap items-center gap-2">
                  <Badge variant="primary">{roleLabel(u.role)}</Badge>
                  {u.nim_nidn && <span className="text-xs text-muted-foreground">{u.nim_nidn}</span>}
                  {u.is_active === false ? <Badge variant="danger">Nonaktif</Badge> : <Badge variant="success">Aktif</Badge>}
                </div>
              </div>
              <UserActions user={u} onToggleActive={onToggleActive} onResetPassword={onResetPassword} />
            </div>
          </CardContent>
        </Card>
      ))}
    </div>
  )
}
