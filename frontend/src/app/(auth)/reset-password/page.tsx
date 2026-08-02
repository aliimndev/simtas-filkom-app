'use client'

import Link from 'next/link'
import { useRouter, useSearchParams } from 'next/navigation'
import { Suspense, useEffect, useMemo, useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useResetPasswordMutation } from '@/lib/hooks/use-auth'
import { getErrorMessage } from '@/lib/utils/error'
import { cn } from '@/lib/utils/cn'

const resetSchema = z
  .object({
    password: z.string().min(8, 'Password minimal 8 karakter'),
    confirm: z.string(),
  })
  .refine((d) => d.password === d.confirm, {
    message: 'Konfirmasi password tidak cocok',
    path: ['confirm'],
  })

type ResetForm = z.infer<typeof resetSchema>

function passwordStrength(pw: string): { score: number; label: string } {
  let score = 0
  if (pw.length >= 8) score++
  if (/[A-Z]/.test(pw)) score++
  if (/[0-9]/.test(pw)) score++
  if (/[^A-Za-z0-9]/.test(pw)) score++
  const labels = ['Lemah', 'Lemah', 'Sedang', 'Kuat', 'Sangat kuat']
  return { score, label: labels[score] }
}

function ResetPasswordInner() {
  const router = useRouter()
  const searchParams = useSearchParams()
  const token = searchParams.get('token') ?? ''
  const [error, setError] = useState<string | null>(null)
  const reset = useResetPasswordMutation()

  const {
    register,
    handleSubmit,
    watch,
    formState: { errors },
  } = useForm<ResetForm>({ resolver: zodResolver(resetSchema) })

  const password = watch('password') ?? ''
  const strength = useMemo(() => passwordStrength(password), [password])

  useEffect(() => {
    if (!token) router.replace('/forgot-password')
  }, [token, router])

  const onSubmit = handleSubmit(async (data) => {
    setError(null)
    try {
      await reset.mutateAsync({ token, new_password: data.password, confirm_password: data.confirm })
      router.push('/login')
    } catch (err) {
      setError(getErrorMessage(err, 'Token tidak valid atau sudah kedaluwarsa.'))
    }
  })

  if (!token) return null

  return (
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-primary-50 via-background to-secondary-50 px-4 py-10">
      <div className="w-full max-w-md">
        <Card>
          <CardHeader className="text-center">
            <CardTitle>Reset Password</CardTitle>
            <CardDescription>Buat password baru untuk akun Anda.</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={onSubmit} noValidate className="space-y-4">
              {error && <Alert variant="danger">{error}</Alert>}

              <div>
                <Label htmlFor="password">Password Baru</Label>
                <Input
                  id="password"
                  type="password"
                  placeholder="Minimal 8 karakter"
                  invalid={!!errors.password}
                  {...register('password')}
                />
                {password.length > 0 && (
                  <div className="mt-2">
                    <div className="flex gap-1">
                      {[0, 1, 2, 3].map((i) => (
                        <div
                          key={i}
                          className={cn(
                            'h-1.5 flex-1 rounded-full transition-colors',
                            i < strength.score
                              ? strength.score >= 3
                                ? 'bg-success'
                                : strength.score === 2
                                  ? 'bg-warning'
                                  : 'bg-danger'
                              : 'bg-muted',
                          )}
                        />
                      ))}
                    </div>
                    <p className="mt-1 text-xs text-muted-foreground">
                      Kekuatan: <span className="font-medium">{strength.label}</span>
                    </p>
                  </div>
                )}
                {errors.password && <p className="mt-1 text-xs text-danger">{errors.password.message}</p>}
              </div>

              <div>
                <Label htmlFor="confirm">Konfirmasi Password</Label>
                <Input
                  id="confirm"
                  type="password"
                  placeholder="Ulangi password baru"
                  invalid={!!errors.confirm}
                  {...register('confirm')}
                />
                {errors.confirm && <p className="mt-1 text-xs text-danger">{errors.confirm.message}</p>}
              </div>

              <Button type="submit" fullWidth loading={reset.isPending}>
                Simpan Password Baru
              </Button>
              <p className="text-center text-sm">
                <Link href="/login" className="font-medium text-primary hover:underline">
                  Kembali ke Login
                </Link>
              </p>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}

export default function ResetPasswordPage() {
  return (
    <Suspense fallback={null}>
      <ResetPasswordInner />
    </Suspense>
  )
}
