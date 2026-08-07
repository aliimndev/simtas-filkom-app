'use client'

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { CalendarPlus, CheckCircle2 } from 'lucide-react'
import { useState } from 'react'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { ListSkeleton } from '@/components/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { academicYearApi } from '@/lib/api/user-api'
import { formatDate } from '@/lib/utils/date'
import { getErrorMessage } from '@/lib/utils/error'
import { Alert } from '@/components/ui/alert'

export default function AcademicYearsPage() {
  const queryClient = useQueryClient()
  const [showForm, setShowForm] = useState(false)
  const [form, setForm] = useState({ name: '', start_date: '', end_date: '' })

  const years = useQuery({ queryKey: ['academic-years'], queryFn: academicYearApi.list })

  const create = useMutation({
    mutationFn: () => academicYearApi.create(form),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['academic-years'] })
      setShowForm(false)
      setForm({ name: '', start_date: '', end_date: '' })
    },
  })

  const activate = useMutation({
    mutationFn: (id: string) => academicYearApi.activate(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['academic-years'] }),
  })

  return (
    <div className="space-y-6">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl font-bold">Tahun Akademik</h1>
          <p className="mt-1 text-sm text-muted-foreground">Kelola periode tahun akademik aktif</p>
        </div>
        <Button onClick={() => setShowForm((v) => !v)}>
          <CalendarPlus className="h-4 w-4" /> Tambah Tahun Akademik
        </Button>
      </div>

      {showForm && (
        <Card>
          <CardContent className="p-6">
            {create.isError && <Alert variant="danger" className="mb-4">{getErrorMessage(create.error)}</Alert>}
            <form
              className="grid grid-cols-1 gap-4 sm:grid-cols-3"
              onSubmit={(e) => {
                e.preventDefault()
                create.mutate()
              }}
            >
              <div>
                <Label htmlFor="name" required>Nama</Label>
                <Input id="name" placeholder="mis. 2026/2027 Ganjil" value={form.name} onChange={(e) => setForm({ ...form, name: e.target.value })} required />
              </div>
              <div>
                <Label htmlFor="start" required>Mulai</Label>
                <Input id="start" type="date" value={form.start_date} onChange={(e) => setForm({ ...form, start_date: e.target.value })} required />
              </div>
              <div>
                <Label htmlFor="end" required>Selesai</Label>
                <Input id="end" type="date" value={form.end_date} onChange={(e) => setForm({ ...form, end_date: e.target.value })} required />
              </div>
              <div className="flex gap-2 sm:col-span-3">
                <Button type="submit" loading={create.isPending}>Simpan</Button>
                <Button type="button" variant="ghost" onClick={() => setShowForm(false)}>Batal</Button>
              </div>
            </form>
          </CardContent>
        </Card>
      )}

      {years.isLoading ? (
        <ListSkeleton count={4} label="Memuat tahun akademik…" />
      ) : (
        <>
          {/* Desktop table */}
          <div className="hidden md:block overflow-x-auto">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>Nama</TableHead>
                  <TableHead>Periode</TableHead>
                  <TableHead>Status</TableHead>
                  <TableHead className="text-right">Aksi</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {years.data?.map((y) => (
                  <TableRow key={y.id}>
                    <TableCell className="font-medium">{y.name}</TableCell>
                    <TableCell>{formatDate(y.start_date)} — {formatDate(y.end_date)}</TableCell>
                    <TableCell>
                      {y.is_active ? <Badge variant="success">Aktif</Badge> : <Badge variant="muted">Tidak aktif</Badge>}
                    </TableCell>
                    <TableCell className="text-right">
                      {!y.is_active && (
                        <Button size="sm" variant="outline" onClick={() => activate.mutate(y.id)}>
                          <CheckCircle2 className="h-3.5 w-3.5" /> Jadikan Aktif
                        </Button>
                      )}
                    </TableCell>
                  </TableRow>
                ))}
              </TableBody>
            </Table>
          </div>

          {/* Mobile cards */}
          <div className="space-y-3 md:hidden">
            {years.data?.map((y) => (
              <Card key={y.id}>
                <CardContent className="p-4">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0 flex-1">
                      <p className="font-medium">{y.name}</p>
                      <p className="mt-1 text-xs text-muted-foreground">
                        {formatDate(y.start_date)} — {formatDate(y.end_date)}
                      </p>
                      <div className="mt-2">
                        {y.is_active ? <Badge variant="success">Aktif</Badge> : <Badge variant="muted">Tidak aktif</Badge>}
                      </div>
                    </div>
                    {!y.is_active && (
                      <Button size="sm" variant="outline" onClick={() => activate.mutate(y.id)}>
                        <CheckCircle2 className="h-3.5 w-3.5" /> Aktifkan
                      </Button>
                    )}
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </>
      )}
    </div>
  )
}
