'use client'

import Link from 'next/link'
import { useState } from 'react'
import { useForm } from 'react-hook-form'
import { zodResolver } from '@hookform/resolvers/zod'
import { z } from 'zod'
import { Alert } from '@/components/ui/alert'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'
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
    <div className="flex min-h-screen items-center justify-center bg-gradient-to-br from-primary-50 via-background to-secondary-50 px-4 py-10">
      <div className="w-full max-w-md">
        <Card>
          <CardHeader className="text-center">
            <CardTitle>Lupa Password</CardTitle>
            <CardDescription>
              Masukkan email terdaftar Anda, kami akan mengirimkan tautan untuk mereset password.
            </CardDescription>
          </CardHeader>
          <CardContent>
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
                  {errors.email && <p className="mt-1 text-xs text-danger">{errors.email.message}</p>}
                </div>
                <Button type="submit" fullWidth loading={forgot.isPending}>
                  Kirim Tautan Reset
                </Button>
                <p className="text-center text-sm">
                  <Link href="/login" className="font-medium text-primary hover:underline">
                    Kembali ke Login
                  </Link>
                </p>
              </form>
            )}
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
