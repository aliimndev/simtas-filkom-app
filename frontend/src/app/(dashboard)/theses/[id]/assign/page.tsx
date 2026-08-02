'use client'

import { useParams, useRouter } from 'next/navigation'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { ArrowLeft, Users } from 'lucide-react'
import Link from 'next/link'
import { useState } from 'react'
import { Alert } from '@/components/ui/alert'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { thesisApi } from '@/lib/api/thesis-api'
import { userApi } from '@/lib/api/user-api'
import { getErrorMessage } from '@/lib/utils/error'
import { formatDate } from '@/lib/utils/date'

export default function AssignSupervisorPage() {
  const params = useParams<{ id: string }>()
  const router = useRouter()
  const queryClient = useQueryClient()
  const id = params.id
  const [selected, setSelected] = useState<string[]>([])
  const [error, setError] = useState<string | null>(null)
  const [q, setQ] = useState('')

  const thesis = useQuery({
    queryKey: ['theses', id],
    queryFn: () => thesisApi.get(id),
    enabled: Boolean(id),
  })

  const lecturers = useQuery({
    queryKey: ['lecturers'],
    queryFn: userApi.lecturers,
  })

  const assign = useMutation({
    mutationFn: () => thesisApi.assignSupervisor(id, { supervisor_ids: selected }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['theses'] })
      router.push('/theses')
    },
    onError: (e) => setError(getErrorMessage(e, 'Gagal menyimpan pembimbing.')),
  })

  const t = thesis.data

  // Sinkronkan pilihan dengan pembimbing yang sudah ada saat data thesis dimuat.
  // Dilakukan saat render (bukan di effect) agar tidak memicu cascading render.
  const [prevSupKey, setPrevSupKey] = useState<string | null>(null)
  const supKey = t?.supervisors?.map((s) => s.id).sort().join(',') ?? ''
  if (supKey !== prevSupKey) {
    setPrevSupKey(supKey)
    setSelected(t?.supervisors?.map((s) => s.id) ?? [])
  }

  function toggle(id: string) {
    setSelected((prev) => (prev.includes(id) ? prev.filter((x) => x !== id) : [...prev, id]))
  }

  return (
    <div className="mx-auto max-w-3xl space-y-6">
      <div>
        <Link href="/theses" className="mb-4 inline-flex items-center gap-1.5 text-sm font-medium text-muted-foreground hover:text-foreground">
          <ArrowLeft className="h-4 w-4" /> Kembali ke Daftar
        </Link>
        <h1 className="text-2xl font-bold">Atur Dosen Pembimbing</h1>
        {t && (
          <p className="mt-1 text-sm text-muted-foreground">
            {t.title} · {t.student?.full_name} · {formatDate(t.submitted_at)}
          </p>
        )}
      </div>

      {error && <Alert variant="danger">{error}</Alert>}

      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2 text-base">
            <Users className="h-5 w-5 text-primary" /> Pilih Pembimbing (maks. 2)
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-4">
          <div>
            <Label htmlFor="search">Cari Dosen</Label>
            <Input id="search" placeholder="Nama dosen…" value={q} onChange={(e) => setQ(e.target.value)} />
          </div>

          <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
            {(lecturers.data ?? []).filter((l) => {
              const term = q.trim().toLowerCase()
              if (!term) return true
              return (
                l.full_name.toLowerCase().includes(term) ||
                (l.nim_nidn ?? '').toLowerCase().includes(term)
              )
            }).map((l) => {
              const isSelected = selected.includes(l.id)
              const isSupervisor = t?.supervisors?.some((s) => s.id === l.id) ?? false
              return (
                <button
                  key={l.id}
                  type="button"
                  onClick={() => {
                    if (isSelected) {
                      toggle(l.id)
                    } else if (selected.length < 2) {
                      toggle(l.id)
                    }
                  }}
                  className={`flex items-center justify-between rounded-lg border p-3 text-left transition-colors ${
                    isSelected ? 'border-primary bg-primary-50' : 'border-border hover:bg-muted'
                  } ${!isSelected && selected.length >= 2 ? 'cursor-not-allowed opacity-50' : ''}`}
                >
                  <div className="min-w-0">
                    <p className="truncate text-sm font-medium">{l.full_name}</p>
                    <p className="truncate text-xs text-muted-foreground">{l.nim_nidn ?? ''}</p>
                  </div>
                  {isSupervisor && <Badge variant="primary">Pembimbing</Badge>}
                  {isSelected && !isSupervisor && <Badge variant="success">Dipilih</Badge>}
                </button>
              )
            })}
          </div>

          <div className="flex gap-3 pt-2">
            <Button onClick={() => assign.mutate()} disabled={selected.length === 0} loading={assign.isPending}>
              Simpan Pembimbing
            </Button>
          </div>
        </CardContent>
      </Card>
    </div>
  )
}
