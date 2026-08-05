'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { ArrowLeft } from 'lucide-react'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { useForgotPasswordMutation } from '@/lib/hooks/use-auth'
import { getErrorMessage } from '@/lib/utils/error'

const forgotSchema = z.object({
  email: z.string().min(1, 'Email wajib diisi').email('Format email tidak valid'),
})

type ForgotForm = z.infer<typeof forgotSchema>

export default function ForgotPasswordPage() {
  const [error, setError] = useState<string | null>(null)
  const [done, setDone] = useState(false)
  const forgot = useForgotPasswordMutation()
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<ForgotForm>({ resolver: zodResolver(forgotSchema) })

  const onSubmit = handleSubmit(async (data) => {
    setError(null)
    try {
      await forgot.mutateAsync(data)
      setDone(true)
    } catch (err) {
      setError(getErrorMessage(err, 'Gagal mengirim email reset. Silakan coba lagi.'))
    }
  })

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4 py-10">
      <div className="w-full max-w-md">
        <div className="st-card rounded-2xl p-8">
          <div className="mb-6 text-center">
            <span className="font-mono text-[0.7rem] uppercase tracking-[0.25em] text-muted-foreground">
              Lupa Password
            </span>
            <h1 className="landing-display mt-3 text-3xl">Atur ulang password</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              Masukkan email terdaftar Anda, kami akan mengirimkan tautan untuk mereset password.
            </p>
          </div>

          {done ? (
            <div className="space-y-4 text-center">
              <Alert variant="success">
                Jika email terdaftar, tautan reset password telah dikirim. Periksa kotak masuk Anda.
              </Alert>
              <Button asChild variant="outline" className="w-full">
                <Link href="/login">Kembali ke Login</Link>
              </Button>
            </div>
          ) : (
            <form onSubmit={onSubmit} noValidate className="space-y-4">
              {error && <Alert variant="danger">{error}</Alert>}
              <div>
                <Label htmlFor="email">Email</Label>
                <Input
                  id="email"
                  type="email"
                  placeholder="nama@unida.ac.id"
                  invalid={!!errors.email}
                  {...register('email')}
                />
                {errors.email && <p className="mt-1 text-xs text-danger-700">{errors.email.message}</p>}
              </div>
              <Button type="submit" fullWidth loading={forgot.isPending}>
                Kirim Tautan Reset
              </Button>
              <p className="text-center text-sm">
                <Link href="/login" className="inline-flex items-center gap-1.5 font-medium text-primary hover:underline">
                  <ArrowLeft className="h-3.5 w-3.5" /> Kembali ke Login
                </Link>
              </p>
            </form>
          )}
        </div>
      </div>
    </div>
  )
}
