'use client'

import { useSearchParams } from 'next/navigation'
import { Suspense, useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { useQueryClient } from '@tanstack/react-query'
import { z } from 'zod'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Textarea } from '@/components/ui/textarea'
import { PasswordInput } from '@/components/ui/password-input'
import { userApi } from '@/lib/api/user-api'
import { useAuthStore } from '@/lib/stores/auth-store'
import { roleLabel } from '@/constants/roles'
import { getErrorMessage } from '@/lib/utils/error'

type ProfileForm = {
  full_name: string
  place_of_birth: string
  address: string
  study_program: string
  nim_nidn: string
  phone: string
  birth_date: string
  faculty: string
  semester: string
}

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

function defaultsFromUser(u: ReturnType<typeof useAuthStore.getState>['user']): ProfileForm {
  return {
    full_name: u?.full_name ?? '',
    place_of_birth: u?.place_of_birth ?? '',
    address: u?.address ?? '',
    study_program: u?.study_program ?? '',
    nim_nidn: u?.nim_nidn ?? '',
    phone: u?.phone ?? '',
    birth_date: u?.birth_date ?? '',
    faculty: u?.faculty ?? '',
    semester: u?.semester != null ? String(u.semester) : '',
  }
}

function ProfileInner() {
  const searchParams = useSearchParams()
  const force = searchParams.get('force') === 'change-password'
  const user = useAuthStore((s) => s.user)
  const setUser = useAuthStore((s) => s.setUser)
  const queryClient = useQueryClient()

  const [pwError, setPwError] = useState<string | null>(null)
  const [pwSuccess, setPwSuccess] = useState(false)
  const [pwSubmitting, setPwSubmitting] = useState(false)

  const [profileError, setProfileError] = useState<string | null>(null)
  const [profileSuccess, setProfileSuccess] = useState(false)
  const [profileSubmitting, setProfileSubmitting] = useState(false)

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<ChangePwForm>({ resolver: zodResolver(changePwSchema) })

  const {
    register: registerProfile,
    handleSubmit: handleProfileSubmit,
    reset: resetProfile,
  } = useForm<ProfileForm>({ defaultValues: defaultsFromUser(user) })

  // Isi ulang form profil begitu data user tersedia / berubah.
  useEffect(() => {
    if (user) resetProfile(defaultsFromUser(user))
  }, [user, resetProfile])

  const onPwSubmit = handleSubmit(async (data) => {
    setPwError(null)
    setPwSuccess(false)
    setPwSubmitting(true)
    try {
      await userApi.changeMyPassword(data.current_password, data.new_password)
      if (user?.must_change_password) setUser({ ...user, must_change_password: false })
      setPwSuccess(true)
      reset()
    } catch (err) {
      setPwError(getErrorMessage(err, 'Gagal mengganti password.'))
    } finally {
      setPwSubmitting(false)
    }
  })

  const onProfileSubmit = handleProfileSubmit(async (data) => {
    setProfileError(null)
    setProfileSuccess(false)
    setProfileSubmitting(true)
    try {
      const payload = {
        full_name: data.full_name,
        nim_nidn: data.nim_nidn || undefined,
        study_program: data.study_program || undefined,
        place_of_birth: data.place_of_birth || undefined,
        address: data.address || undefined,
        phone: data.phone || undefined,
        birth_date: data.birth_date || undefined,
        faculty: data.faculty || undefined,
        semester: data.semester ? Number(data.semester) : undefined,
      }
      const updated = await userApi.updateMe(payload)
      setUser(updated)
      queryClient.invalidateQueries({ queryKey: ['auth', 'me'] })
      setProfileSuccess(true)
    } catch (err) {
      setProfileError(getErrorMessage(err, 'Gagal menyimpan profil.'))
    } finally {
      setProfileSubmitting(false)
    }
  })

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      {force && (
        <Alert variant="warning">
          Ini adalah login pertama Anda. Harap ganti password default sebelum melanjutkan.
        </Alert>
      )}

      {/* Header Profil */}
      <Card className="overflow-hidden">
        <div className="bg-(--st-accent-from)/10 px-6 py-5">
          <p className="landing-eyebrow">Profil</p>
          <h1 className="mt-1 text-xl font-semibold text-st-text">{user?.full_name ?? '—'}</h1>
          <p className="mt-1 text-sm text-st-muted">
            {roleLabel(user?.role)} · NIM {user?.nim_nidn ?? '—'} · {user?.study_program ?? '—'}
          </p>
        </div>
      </Card>

      {/* Informasi Pribadi */}
      <Card>
        <CardHeader>
          <p className="landing-eyebrow">Data Diri</p>
          <CardTitle className="mt-1 landing-heading text-xl">Informasi <span className="accent-text italic">Pribadi</span></CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onProfileSubmit} noValidate className="space-y-4">
            {profileError && <Alert variant="danger">{profileError}</Alert>}
            {profileSuccess && <Alert variant="success">Profil berhasil disimpan.</Alert>}
            <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <Label htmlFor="full_name" required>Nama Mahasiswa</Label>
                <Input id="full_name" {...registerProfile('full_name')} />
              </div>
              <div>
                <Label htmlFor="email">Email</Label>
                <Input id="email" value={user?.email ?? ''} disabled />
              </div>
              <div>
                <Label htmlFor="place_of_birth">Tempat Lahir</Label>
                <Input id="place_of_birth" {...registerProfile('place_of_birth')} />
              </div>
              <div>
                <Label htmlFor="study_program">Jurusan</Label>
                <Input id="study_program" {...registerProfile('study_program')} />
              </div>
            </div>
            <div>
              <Label htmlFor="address">Alamat Rumah</Label>
              <Textarea id="address" {...registerProfile('address')} />
            </div>

            {/* Informasi Akademik */}
            <div className="border-t border-st-stroke pt-4">
              <p className="mb-3 text-sm font-medium text-st-text">Informasi Akademik</p>
              <div className="grid grid-cols-1 gap-4 sm:grid-cols-2">
                <div>
                  <Label htmlFor="nim_nidn">NIM</Label>
                  <Input id="nim_nidn" {...registerProfile('nim_nidn')} />
                </div>
                <div>
                  <Label htmlFor="phone">Nomor Telepon</Label>
                  <Input id="phone" type="tel" {...registerProfile('phone')} />
                </div>
                <div>
                  <Label htmlFor="birth_date">Tanggal Lahir</Label>
                  <Input id="birth_date" type="date" {...registerProfile('birth_date')} />
                </div>
                <div>
                  <Label htmlFor="faculty">Fakultas</Label>
                  <Input id="faculty" {...registerProfile('faculty')} />
                </div>
                <div>
                  <Label htmlFor="semester">Semester</Label>
                  <Input id="semester" type="number" min={1} {...registerProfile('semester')} />
                </div>
              </div>
            </div>

            <Button type="submit" loading={profileSubmitting}>Simpan</Button>
          </form>
        </CardContent>
      </Card>

      <Card>
        <CardHeader>
          <p className="landing-eyebrow">Keamanan</p>
          <CardTitle className="mt-1 landing-heading text-xl">Ubah <span className="accent-text italic">Password</span></CardTitle>
        </CardHeader>
        <CardContent>
          <form onSubmit={onPwSubmit} noValidate className="space-y-4">
            {pwError && <Alert variant="danger">{pwError}</Alert>}
            {pwSuccess && <Alert variant="success">Password berhasil diubah.</Alert>}
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
            <Button type="submit" loading={pwSubmitting}>Simpan Password</Button>
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
