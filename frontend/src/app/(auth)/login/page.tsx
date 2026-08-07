'use client'

import Link from 'next/link'
import { useRouter } from 'next/navigation'
import { Suspense, useEffect, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { ArrowRight } from 'lucide-react'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { PasswordInput } from '@/components/ui/password-input'
import { useLoginMutation } from '@/lib/hooks/use-auth'
import { useAuthStore } from '@/lib/stores/auth-store'
import { getErrorMessage, mapAuthError } from '@/lib/utils/error'

const loginSchema = z.object({
  email: z.string().min(1, 'Email wajib diisi').email('Format email tidak valid'),
  password: z.string().min(1, 'Password wajib diisi'),
})

type LoginForm = z.infer<typeof loginSchema>

function BrandPanel() {
  return (
    <div className="relative hidden overflow-hidden border-r border-border bg-background lg:flex lg:flex-col lg:justify-between lg:p-12">
      <div
        aria-hidden
        className="pointer-events-none absolute left-1/2 top-[-15%] h-128 w-lg -translate-x-1/2 rounded-full opacity-25 blur-[120px]"
        style={{ background: 'radial-gradient(circle, var(--st-accent-to), transparent 60%)' }}
      />
      <div
        aria-hidden
        className="landing-grid-bg pointer-events-none absolute inset-0 opacity-60"
      />
      <Link href="/" className="accent-ring relative flex w-fit items-center gap-2.5">
        <span className="flex h-9 w-9 items-center justify-center rounded-full bg-background">
          <span className="font-display text-[15px] italic text-foreground">sf</span>
        </span>
        <span className="text-sm font-medium tracking-tight text-foreground">
          SIMTAS <span className="text-muted-foreground">FILKOM</span>
        </span>
      </Link>

      <div className="relative max-w-md">
        <p className="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-muted-foreground">
          SIMTAS://FILKOM
        </p>
        <h1 className="landing-display mt-5 text-4xl md:text-5xl">
          Satu sistem untuk seluruh perjalanan{' '}
          <span className="accent-text italic">Tugas Akhir Skripsi</span>.
        </h1>
        <p className="mt-5 text-sm leading-relaxed text-muted-foreground md:text-base">
          Dari pengajuan judul, bimbingan, seminar, hingga sidang dan arsip—seluruh
          proses Tugas Akhir Fakultas Ilmu Komputer dalam satu ekosistem digital.
        </p>
      </div>

      <p className="relative font-mono text-[0.7rem] uppercase tracking-[0.2em] text-muted-foreground">
        Fakultas Ilmu Komputer · Universitas Djuanda
      </p>
    </div>
  )
}

function LoginForm() {
  const router = useRouter()
  const accessToken = useAuthStore((s) => s.accessToken)
  const [error, setError] = useState<string | null>(null)
  const login = useLoginMutation()

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
    <div className="flex w-full flex-col justify-center px-6 py-10 sm:px-10">
      <div className="mx-auto w-full max-w-sm">
        <div className="mb-8">
          <span className="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-muted-foreground">
            Masuk ke Sistem
          </span>
          <h2 className="landing-display mt-3 text-3xl md:text-4xl">Selamat datang kembali.</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            Gunakan kredensial yang diberikan administrator fakultas.
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
              placeholder="nama@unida.ac.id"
              autoComplete="email"
              invalid={!!errors.email}
              {...register('email')}
            />
            {errors.email && <p className="mt-1 text-xs text-danger-700">{errors.email.message}</p>}
          </div>

          <div>
            <div className="flex items-center justify-between">
              <Label htmlFor="password">Password</Label>
              <Link
                href="/forgot-password"
                className="mb-1.5 text-xs font-medium text-primary hover:underline"
              >
                Lupa password?
              </Link>
            </div>
            <PasswordInput
              id="password"
              placeholder="••••••••"
              autoComplete="current-password"
              invalid={!!errors.password}
              {...register('password')}
            />
            {errors.password && <p className="mt-1 text-xs text-danger-700">{errors.password.message}</p>}
          </div>

          <Button type="submit" fullWidth size="lg" loading={login.isPending}>
            Masuk <ArrowRight className="h-4 w-4" />
          </Button>
        </form>

        <p className="mt-8 text-center font-mono text-[0.7rem] uppercase tracking-[0.2em] text-muted-foreground">
          FILKOM Universitas Djuanda — © {new Date().getFullYear()}
        </p>
      </div>
    </div>
  )
}

export default function LoginPage() {
  return (
    <div className="grid min-h-screen lg:grid-cols-2">
      <BrandPanel />
      <Suspense fallback={null}>
        <LoginForm />
      </Suspense>
    </div>
  )
}
