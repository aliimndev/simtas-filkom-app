'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { UserPlus, Search, KeyRound, Power, PowerOff } from 'lucide-react'
import { useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { ListSkeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { userApi } from '@/lib/api/user-api'
import { roleLabel, ROLE_OPTIONS } from '@/constants/roles'
import { getErrorMessage } from '@/lib/utils/error'
import { Alert } from '@/components/ui/alert'
import type { User } from '@/types/auth'

export default function AdminUsersPage() {
  const queryClient = useQueryClient()
  const [q, setQ] = useState('')
  const [role, setRole] = useState('')
  const [page, setPage] = useState(1)
  const [actionError, setActionError] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)
  const [createForm, setCreateForm] = useState({ email: '', full_name: '', role: 'mahasiswa', nim_nidn: '' })

  const users = useQuery({
    queryKey: ['admin', 'users', q, role, page],
    queryFn: () => userApi.list({ q: q || undefined, role: role || undefined, page, per_page: 10 }),
  })

  const list = users.data?.data ?? []
  const total = users.data?.meta.total ?? 0

  const toggleActive = useMutation({
    mutationFn: (u: User) => (u.is_active === false ? userApi.activate(u.id) : userApi.deactivate(u.id)),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
      setActionError(null)
    },
  })

  const create = useMutation({
    mutationFn: (data: typeof createForm) => userApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
      setShowCreate(false)
      setCreateForm({ email: '', full_name: '', role: 'mahasiswa', nim_nidn: '' })
    },
  })

  const resetPassword = useMutation({
    mutationFn: (id: string) => userApi.resetPassword(id),
    onSuccess: () => setActionError(null),
    onError: (e) => setActionError(getErrorMessage(e, 'Gagal mereset password.')),
  })

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">Manajemen Pengguna</h1>
          <p className="mt-1 text-sm text-muted-foreground">Kelola akun mahasiswa, dosen, dan staf</p>
        </div>
        <Button onClick={() => setShowCreate((v) => !v)}>
          <UserPlus className="h-4 w-4" /> Tambah Pengguna
        </Button>
      </div>

      {showCreate && (
        <Card>
          <CardContent className="p-6">
            <h2 className="mb-4 text-base font-semibold">Tambah Pengguna Baru</h2>
            {create.isError && <Alert variant="danger" className="mb-4">{getErrorMessage(create.error)}</Alert>}
            <form
              className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"
              onSubmit={(e) => {
                e.preventDefault()
                create.mutate(createForm)
              }}
            >
              <Input placeholder="Nama lengkap" value={createForm.full_name} onChange={(e) => setCreateForm({ ...createForm, full_name: e.target.value })} required />
              <Input placeholder="Email" type="email" value={createForm.email} onChange={(e) => setCreateForm({ ...createForm, email: e.target.value })} required />
              <Input placeholder="NIM / NIDN" value={createForm.nim_nidn} onChange={(e) => setCreateForm({ ...createForm, nim_nidn: e.target.value })} />
              <Select value={createForm.role} onChange={(e) => setCreateForm({ ...createForm, role: e.target.value })}>
                {ROLE_OPTIONS.map((r) => (
                  <option key={r.value} value={r.value}>{r.label}</option>
                ))}
              </Select>
              <p className="text-xs text-muted-foreground sm:col-span-2 lg:col-span-3">
                Password sementara akan dibuat otomatis dan dikirim ke email pengguna.
              </p>
              <div className="flex gap-2">
                <Button type="submit" loading={create.isPending}>Simpan</Button>
                <Button type="button" variant="ghost" onClick={() => setShowCreate(false)}>Batal</Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      <div className="flex flex-wrap items-center gap-3">
        <div className="relative">
          <Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            placeholder="Cari nama / email / NIM…"
            value={q}
            onChange={(e) => { setQ(e.target.value); setPage(1) }}
            className="max-w-xs pl-9"
          />
        </div>
        <Select value={role} onChange={(e) => { setRole(e.target.value); setPage(1) }} className="w-44">
          <option value="">Semua peran</option>
          {ROLE_OPTIONS.map((r) => (
            <option key={r.value} value={r.value}>{r.label}</option>
          ))}
        </Select>
      </div>

      {actionError && <Alert variant="danger">{actionError}</Alert>}

      {users.isLoading ? (
        <ListSkeleton count={5} label="Memuat pengguna…" />
      ) : (
        <>
          {/* Desktop table */}
          <div className="hidden md:block overflow-x-auto">
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
                {list.map((u) => (
                  <TableRow key={u.id}>
                    <TableCell className="font-medium">{u.full_name}</TableCell>
                    <TableCell>{u.email}</TableCell>
                    <TableCell><Badge variant="primary">{roleLabel(u.role)}</Badge></TableCell>
                    <TableCell>{u.nim_nidn ?? '—'}</TableCell>
                    <TableCell>
                      {u.is_active === false ? <Badge variant="danger">Nonaktif</Badge> : <Badge variant="success">Aktif</Badge>}
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex justify-end gap-1">
                        <Button
                          size="sm"
                          variant="ghost"
                          title={u.is_active === false ? 'Aktifkan' : 'Nonaktifkan'}
                          onClick={() => toggleActive.mutate(u)}
                        >
                          {u.is_active === false ? <Power className="h-4 w-4 text-success" /> : <PowerOff className="h-4 w-4 text-danger-700" />}
                        </Button>
                        <Button
                          size="sm"
                          variant="ghost"
                          title="Reset password"
                          onClick={() => {
                            if (window.confirm(`Reset password untuk ${u.full_name}? Password baru akan dikirim ke email.`)) {
                              resetPassword.mutate(u.id)
                            }
                          }}
                        >
                          <KeyRound className="h-4 w-4" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Mobile cards */}
          <div className="space-y-3 md:hidden">
            {list.map((u) => (
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
                    <div className="flex flex-col gap-1.5">
                      <Button
                        size="sm"
                        variant="ghost"
                        title={u.is_active === false ? 'Aktifkan' : 'Nonaktifkan'}
                        onClick={() => toggleActive.mutate(u)}
                      >
                        {u.is_active === false ? <Power className="h-4 w-4 text-success" /> : <PowerOff className="h-4 w-4 text-danger-700" />}
                      </Button>
                      <Button
                        size="sm"
                        variant="ghost"
                        title="Reset password"
                        onClick={() => {
                          if (window.confirm(`Reset password untuk ${u.full_name}? Password baru akan dikirim ke email.`)) {
                            resetPassword.mutate(u.id)
                          }
                        }}
                      >
                        <KeyRound className="h-4 w-4" />
                      </Button>
                    </div>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>

          {total > 10 && (
            <div className="flex items-center justify-between">
              <p className="text-sm text-muted-foreground">Total {total} pengguna · Halaman {page}</p>
              <div className="flex gap-2">
                <Button size="sm" variant="outline" disabled={page <= 1} onClick={() => setPage((p) => p - 1)}>Sebelumnya</Button>
                <Button size="sm" variant="outline" disabled={page >= Math.ceil(total / 10)} onClick={() => setPage((p) => p + 1)}>Berikutnya</Button>
              </div>
            </div>
          )}
        </>
      )}
    </div>
  )
}
