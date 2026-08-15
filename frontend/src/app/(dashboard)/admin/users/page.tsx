'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { UserPlus } from 'lucide-react'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { ListSkeleton } from '@/components/ui/skeleton'
import { Alert } from '@/components/ui/alert'
import { userApi } from '@/lib/api/user-api'
import { getErrorMessage } from '@/lib/utils/error'
import { UserFilters } from '@/components/features/admin-users/user-filters'
import { UserTable } from '@/components/features/admin-users/user-table'
import { UserMobileCards } from '@/components/features/admin-users/user-mobile-cards'
import { CreateUserForm, type CreateUserFormData } from '@/components/features/admin-users/create-user-form'
import { Pagination } from '@/components/features/admin-users/pagination'
import type { User } from '@/types/auth'

const PAGE_SIZE = 10

export default function AdminUsersPage() {
  const queryClient = useQueryClient()
  const [q, setQ] = useState('')
  const [role, setRole] = useState('')
  const [page, setPage] = useState(1)
  const [actionError, setActionError] = useState<string | null>(null)
  const [showCreate, setShowCreate] = useState(false)

  const users = useQuery({
    queryKey: ['admin', 'users', q, role, page],
    queryFn: () => userApi.list({ q: q || undefined, role: role || undefined, page, per_page: PAGE_SIZE }),
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
    mutationFn: (data: CreateUserFormData) => userApi.create(data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin', 'users'] })
      setShowCreate(false)
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
          <p className="landing-eyebrow">Manajemen Pengguna</p>
          <h1 className="mt-2 text-balance landing-heading text-2xl">
            Kelola akun <span className="accent-text italic">mahasiswa</span>, dosen, dan staf
          </h1>
        </div>
        <Button onClick={() => setShowCreate((v) => !v)}>
          <UserPlus className="h-4 w-4" /> Tambah Pengguna
        </Button>
      </div>

      <CreateUserForm
        isOpen={showCreate}
        onClose={() => setShowCreate(false)}
        isPending={create.isPending}
        isError={create.isError}
        errorMessage={create.isError ? getErrorMessage(create.error) : null}
        onSubmit={(data) => create.mutate(data)}
      />

      <UserFilters
        q={q}
        role={role}
        onQChange={(next) => { setQ(next); setPage(1) }}
        onRoleChange={(next) => { setRole(next); setPage(1) }}
      />

      {actionError && <Alert variant="danger">{actionError}</Alert>}

      {users.isLoading ? (
        <ListSkeleton count={5} label="Memuat pengguna…" />
      ) : (
        <>
          <UserTable
            users={list}
            onToggleActive={(u) => toggleActive.mutate(u)}
            onResetPassword={(u) => {
              if (window.confirm(`Reset password untuk ${u.full_name}? Password baru akan dikirim ke email.`)) {
                resetPassword.mutate(u.id)
              }
            }}
          />
          <UserMobileCards
            users={list}
            onToggleActive={(u) => toggleActive.mutate(u)}
            onResetPassword={(u) => {
              if (window.confirm(`Reset password untuk ${u.full_name}? Password baru akan dikirim ke email.`)) {
                resetPassword.mutate(u.id)
              }
            }}
          />
          <Pagination page={page} total={total} perPage={PAGE_SIZE} onPageChange={setPage} />
        </>
      )}
    </div>
  )
}
