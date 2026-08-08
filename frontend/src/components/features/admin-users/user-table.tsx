import { Badge } from '@/components/ui/badge'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { roleLabel } from '@/constants/roles'
import { UserActions } from './user-actions'
import type { User } from '@/types/auth'

/** Tabel pengguna (desktop). */
export function UserTable({
  users,
  onToggleActive,
  onResetPassword,
}: {
  users: User[]
  onToggleActive: (user: User) => void
  onResetPassword: (user: User) => void
}) {
  return (
    <div className="hidden overflow-x-auto md:block">
      <Table>
        <TableHeader>
          <TableRow>
            <TableHead>Nama</TableHead>
            <TableHead>Email</TableHead>
            <TableHead>Peran</TableHead>
            <TableHead>NIM/NIDN</TableHead>
            <TableHead>Status</TableHead>
            <TableHead className="text-right">Aksi</TableHead>
          </TableRow>
        </TableHeader>
        <TableBody>
          {users.map((u) => (
            <TableRow key={u.id}>
              <TableCell className="font-medium">{u.full_name}</TableCell>
              <TableCell>{u.email}</TableCell>
              <TableCell><Badge variant="primary">{roleLabel(u.role)}</Badge></TableCell>
              <TableCell>{u.nim_nidn ?? '—'}</TableCell>
              <TableCell>
                {u.is_active === false ? <Badge variant="danger">Nonaktif</Badge> : <Badge variant="success">Aktif</Badge>}
              </TableCell>
              <TableCell className="text-right">
                <UserActions user={u} layout="row" onToggleActive={onToggleActive} onResetPassword={onResetPassword} />
              </TableCell>
            </TableRow>
          ))}
        </TableBody>
      </Table>
    </div>
  )
}
