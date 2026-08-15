'use client'

import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Select } from '@/components/ui/select'
import { ListSkeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
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
      <div>
        <p className="landing-eyebrow">Audit Log</p>
        <h1 className="mt-2 text-balance landing-heading text-2xl">
          Jejak aktivitas seluruh <span className="accent-text italic">pengguna</span> sistem
        </h1>
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
        <ListSkeleton count={6} label="Memuat audit log…" />
      ) : (
        <>
          {/* Desktop table */}
          <div className="hidden md:block overflow-x-auto">
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
          </div>

          {/* Mobile cards */}
          <div className="space-y-3 md:hidden">
            {list.map((l) => (
              <Card key={l.id}>
                <CardContent className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <div className="flex items-center gap-2">
                        <span className="whitespace-nowrap text-xs text-muted-foreground">
                          {formatDateTime(l.created_at)}
                        </span>
                        <Badge variant="muted" className="font-mono text-[10px]">{l.action}</Badge>
                      </div>
                      <p className="mt-1.5 text-sm font-medium">{l.user?.full_name ?? '—'}</p>
                      <p className="mt-0.5 text-xs text-muted-foreground">
                        {l.entity_type ?? '—'}
                        {l.entity_id ? ` · ${l.entity_id.slice(0, 8)}…` : ''}
                      </p>
                    </div>
                    <span className="shrink-0 text-xs text-muted-foreground">{l.ip_address ?? '—'}</span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

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
