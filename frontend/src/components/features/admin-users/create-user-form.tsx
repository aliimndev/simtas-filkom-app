'use client'

import { useState } from 'react'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Select } from '@/components/ui/select'
import { ROLE_OPTIONS } from '@/constants/roles'

export interface CreateUserFormData {
  email: string
  full_name: string
  role: string
  nim_nidn: string
}

/** Kartu "Tambah Pengguna Baru" — state form hidup di komponen ini. */
export function CreateUserForm({
  isOpen,
  onClose,
  isPending,
  isError,
  errorMessage,
  onSubmit,
}: {
  isOpen: boolean
  onClose: () => void
  isPending: boolean
  isError: boolean
  errorMessage: string | null
  onSubmit: (data: CreateUserFormData) => void
}) {
  const [form, setForm] = useState<CreateUserFormData>({
    email: '',
    full_name: '',
    role: 'mahasiswa',
    nim_nidn: '',
  })

  if (!isOpen) return null

  return (
    <Card>
      <CardContent className="p-6">
        <h2 className="mb-4 text-base font-semibold">Tambah Pengguna Baru</h2>
        {isError && <Alert variant="danger" className="mb-4">{errorMessage}</Alert>}
        <form
          className="grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-3"
          onSubmit={(e) => {
            e.preventDefault()
            onSubmit(form)
          }}
        >
          <Input placeholder="Nama lengkap" value={form.full_name} onChange={(e) => setForm({ ...form, full_name: e.target.value })} required />
          <Input placeholder="Email" type="email" value={form.email} onChange={(e) => setForm({ ...form, email: e.target.value })} required />
          <Input placeholder="NIM / NIDN" value={form.nim_nidn} onChange={(e) => setForm({ ...form, nim_nidn: e.target.value })} />
          <Select value={form.role} onChange={(e) => setForm({ ...form, role: e.target.value })}>
            {ROLE_OPTIONS.map((r) => (
              <option key={r.value} value={r.value}>{r.label}</option>
            ))}
          </Select>
          <p className="text-xs text-muted-foreground sm:col-span-2 lg:col-span-3">
            Password sementara akan dibuat otomatis dan dikirim ke email pengguna.
          </p>
          <div className="flex gap-2">
            <Button type="submit" loading={isPending}>Simpan</Button>
            <Button type="button" variant="ghost" onClick={onClose}>Batal</Button>
          </div>
        </form>
      </CardContent>
    </Card>
  )
}
