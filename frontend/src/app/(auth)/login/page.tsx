'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Suspense, useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useLoginMutation } from '@/lib/hooks/use-auth'
import { useAuthStore } from '@/lib/stores/auth-store'
import { getErrorMessage, mapAuthError } from '@/lib/utils/error'

const loginSchema = z.object({
  email: z.string().min(1, 'Email wajib diisi').email('Format email tidak valid'),
  password: z.string().min(1, 'Password wajib diisi'),
})

type LoginForm = z.infer<typeof loginSchema>

function LoginForm() {
  const router = useRouter()
  const accessToken = useAuthStore((s) => s.accessToken)
  const [error, setError] = useState<string | null>(null)
  const login = useLoginMutation()

  // Doc 15: jika sudah login, arahkan langsung ke dashboard.
  useEffect(() => {
    if (accessToken) router.replace('/dashboard')
  }, [accessToken, router])
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginForm>({ resolver: zodResolver(loginSchema) })

  const onSubmit = handleSubmit(async (data) => {
    setError(null)
    try {
      await login.mutateAsync(data)
    } catch (err) {
      setError(mapAuthError(getErrorMessage(err, 'Login gagal. Periksa email dan password Anda.')))
    }
  })

  return (
    <div className="w-full max-w-md">
      <div className="rounded-2xl border border-border bg-background p-8 shadow-xl">
        <div className="mb-8 text-center">
          <div className="mx-auto mb-4 flex h-16 w-16 items-center justify-center rounded-2xl bg-primary text-2xl font-black text-primary-foreground shadow-lg">
            SF
          </div>
          <h1 className="text-2xl font-bold tracking-tight">SIMTAS FILKOM</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            Sistem Manajemen Tugas Akhir & Skripsi
          </p>
        </div>

        {error && (
          <Alert variant="danger" className="mb-4">
            {error}
          </Alert>
        )}

        <form onSubmit={onSubmit} noValidate className="space-y-4">
          <div>
            <Label htmlFor="email">Email</Label>
            <Input
              id="email"
              type="email"
              placeholder="nama@student.unida.ac.id"
              autoComplete="email"
              invalid={!!errors.email}
              {...register('email')}
            />
            {errors.email && <p className="mt-1 text-xs text-danger">{errors.email.message}</p>}
          </div>

          <div>
            <div className="flex items-center justify-between">
              <Label htmlFor="password">Password</Label>
              <Link
                href="/forgot-password"
                className="mb-1.5 text-xs font-medium text-primary hover:text-primary-700 hover:underline"
              >
                Lupa password?
              </Link>
            </div>
            <Input
              id="password"
              type="password"
              placeholder="••••••••"
              autoComplete="current-password"
              invalid={!!errors.password}
              {...register('password')}
            />
            {errors.password && <p className="mt-1 text-xs text-danger">{errors.password.message}</p>}
          </div>

          <Button type="submit" fullWidth size="lg" loading={login.isPending}>
            Masuk
          </Button>
        </form>
      </div>

      <p className="mt-6 text-center text-xs text-muted-foreground">
        FILKOM Universitas Djuanda — © {new Date().getFullYear()}
      </p>
    </div>
  )
}

export default function LoginPage() {
  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-primary-50 via-background to-secondary-50 px-4 py-10">
      <Suspense fallback={null}>
        <LoginForm />
      </Suspense>
    </div>
  )
}
