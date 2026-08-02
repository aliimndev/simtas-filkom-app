'use client'

import { useQuery } from '@tanstack/react-query'
import { ShieldCheck } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'
import { Spinner } from '@/components/ui/spinner'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import apiClient from '@/lib/api/client'
import { formatDateTime } from '@/lib/utils/date'
import type { PaginatedResponse } from '@/types/api'

interface AuditLog {
  id: string
  user_id?: string
  user?: { full_name: string; email: string } | null
  action: string
  entity_type?: string
  entity_id?: string
  ip_address?: string
  user_agent?: string
  created_at: string
}

export default function AuditLogsPage() {
  const [action, setAction] = useState('')
  const [page, setPage] = useState(1)

  const logs = useQuery({
    queryKey: ['admin', 'audit-logs', action, page],
    queryFn: async () => {
      const res = await apiClient.get<PaginatedResponse<AuditLog>>('/admin/audit-logs', {
        params: { action: action || undefined, page, per_page: 25 },
      })
      return res.data
    },
  })

  const list = logs.data?.data ?? []
  const total = logs.data?.meta.total ?? 0

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-3">
        <div className="flex h-10 w-10 items-center justify-center rounded-xl bg-primary-50 text-primary">
          <ShieldCheck className="h-5 w-5" />
        </div>
        <div>
          <h1 className="text-2xl font-bold">Audit Log</h1>
          <p className="text-sm text-muted-foreground">Jejak aktivitas seluruh pengguna sistem</p>
        </div>
      </div>

      <div className="flex flex-wrap items-center gap-3">
        <Select value={action} onChange={(e) => { setAction(e.target.value); setPage(1) }} className="w-56">
          <option value="">Semua aksi</option>
          <option value="USER_LOGIN">Login</option>
          <option value="USER_LOGOUT">Logout</option>
          <option value="USER_CREATED">User dibuat</option>
          <option value="THESIS_SUBMITTED">Skripsi diajukan</option>
          <option value="THESIS_REVIEWED">Skripsi direview</option>
          <option value="DOCUMENT_UPLOADED">Dokumen diunggah</option>
          <option value="SEMINAR_SCHEDULED">Seminar dijadwalkan</option>
          <option value="DEFENSE_SCHEDULED">Sidang dijadwalkan</option>
          <option value="THESIS_GRADUATED">Wisuda</option>
        </Select>
      </div>

      {logs.isLoading ? (
        <Spinner />
      ) : (
        <>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>Waktu</TableHead>
                <TableHead>Pengguna</TableHead>
                <TableHead>Aksi</TableHead>
                <TableHead>Entitas</TableHead>
                <TableHead>IP</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {list.map((l) => (
                <TableRow key={l.id}>
                  <TableCell className="whitespace-nowrap text-xs">{formatDateTime(l.created_at)}</TableCell>
                  <TableCell>{l.user?.full_name ?? '—'}</TableCell>
                  <TableCell><Badge variant="muted" className="font-mono">{l.action}</Badge></TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {l.entity_type ?? '—'}
                    {l.entity_id ? ` · ${l.entity_id.slice(0, 8)}…` : ''}
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">{l.ip_address ?? '—'}</TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>

          {total > 25 && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">Total {total} log · Halaman {page}</p>
              <div className="flex gap-2">
                <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Sebelumnya</Button>
                <Button size="sm" variant="outline" disabled={page >= Math.ceil(total / 25)} onClick={() => setPage((p) => p + 1)}>Berikutnya</Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
