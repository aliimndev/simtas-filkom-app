'use client'

import { useSearchParams } from 'next/navigation'
import { Suspense, useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Label } from '@/components/ui/label'
import { PasswordInput } from '@/components/ui/password-input'
import { userApi } from '@/lib/api/user-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { getErrorMessage } from '@/lib/utils/error'

const changePwSchema = z
  .object({
    current_password: z.string().min(1, 'Password saat ini wajib diisi'),
    new_password: z.string().min(8, 'Password minimal 8 karakter'),
    confirm: z.string(),
  })
  .refine((d) => d.new_password === d.confirm, {
    message: 'Konfirmasi password tidak cocok',
    path: ['confirm'],
  })

type ChangePwForm = z.infer<typeof changePwSchema>

function ProfileInner() {
  const searchParams = useSearchParams()
  const force = searchParams.get('force') === 'change-password'
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const [error, setError] = useState<string | null>(null)
  const [success, setSuccess] = useState(false)
  const [submitting, setSubmitting] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ChangePwForm>({ resolver: zodResolver(changePwSchema) })

  const onSubmit = handleSubmit(async (data) => {
    setError(null)
    setSuccess(false)
    setSubmitting(true)
    try {
      await userApi.changeMyPassword(data.current_password, data.new_password)
      if (user?.must_change_password) setUser({ ...user, must_change_password: false })
      setSuccess(true)
      reset()
    } catch (err) {
      setError(getErrorMessage(err, 'Gagal mengganti password.'))
    } finally {
      setSubmitting(false)
    }
  })

  useEffect(() => {
    if (!force) return
    // Fokus otomatis ke form ganti password saat dipaksa
    document.getElementById('current_password')?.focus()
  }, [force])

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {force && (
        <Alert variant="warning">
          Ini adalah login pertama Anda. Harap ganti password default sebelum melanjutkan.
        </Alert>
      )}

      <Card>
        <CardHeader>
          <p className="landing-eyebrow">Profil</p>
          <CardTitle className="mt-1 landing-heading text-xl">Informasi <span className="accent-text italic">Akun</span> Anda</CardTitle>
        </CardHeader>
        <CardContent className="grid grid-cols-1 gap-4 sm:grid-cols-2">
          <div>
            <p className="text-sm text-muted-foreground">Nama Lengkap</p>
            <p className="font-medium">{user?.full_name ?? '—'}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Email</p>
            <p className="font-medium">{user?.email ?? '—'}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">Peran</p>
            <p className="font-medium">{roleLabel(user?.role)}</p>
          </div>
          <div>
            <p className="text-sm text-muted-foreground">NIM / NIDN</p>
            <p className="font-medium">{user?.nim_nidn ?? '—'}</p>
          </div>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <p className="landing-eyebrow">Keamanan</p>
          <CardTitle className="mt-1 landing-heading text-xl">Ubah <span className="accent-text italic">Password</span></CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onSubmit} noValidate className="space-y-4">
            {error && <Alert variant="danger">{error}</Alert>}
            {success && <Alert variant="success">Password berhasil diubah.</Alert>}
            <div>
              <Label htmlFor="current_password" required>Password Saat Ini</Label>
              <PasswordInput id="current_password" autoComplete="current-password" invalid={!!errors.current_password} {...register('current_password')} />
              {errors.current_password && <p className="mt-1 text-xs text-danger">{errors.current_password.message}</p>}
            </div>
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <Label htmlFor="new_password" required>Password Baru</Label>
                <PasswordInput id="new_password" autoComplete="new-password" invalid={!!errors.new_password} {...register('new_password')} />
                {errors.new_password && <p className="mt-1 text-xs text-danger">{errors.new_password.message}</p>}
              </div>
              <div>
                <Label htmlFor="confirm" required>Konfirmasi</Label>
                <PasswordInput id="confirm" autoComplete="new-password" invalid={!!errors.confirm} {...register('confirm')} />
                {errors.confirm && <p className="mt-1 text-xs text-danger">{errors.confirm.message}</p>}
              </div>
            </div>
            <Button type="submit" loading={submitting}>Simpan Password</Button>
          </form>
        </CardContent>
      </Card>
    </div>
  )
}

export default function ProfilePage() {
  return (
    <Suspense fallback={null}>
      <ProfileInner />
    </Suspense>
  )
}
