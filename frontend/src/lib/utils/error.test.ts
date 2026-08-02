import { AxiosError, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios'
import { getErrorMessage, mapAuthError } from './error'

function makeAxiosError(
  status: number | undefined,
  data: unknown,
  code?: string,
): AxiosError {
  const response = status === undefined ? undefined : ({ status, data } as AxiosResponse)
  return new AxiosError('Request failed', code, {} as InternalAxiosRequestConfig, undefined, response)
}

describe('getErrorMessage', () => {
  it('returns backend message when present', () => {
    const err = makeAxiosError(400, { message: 'Email sudah terdaftar' })
    expect(getErrorMessage(err)).toBe('Email sudah terdaftar')
  })

  it('returns first field error from errors map', () => {
    const err = makeAxiosError(422, { errors: { email: 'Format email tidak valid' } })
    expect(getErrorMessage(err)).toBe('Format email tidak valid')
  })

  it('returns error string body when message missing', () => {
    const err = makeAxiosError(500, { error: 'Internal server error' })
    expect(getErrorMessage(err)).toBe('Internal server error')
  })

  it('maps timeout code', () => {
    const err = makeAxiosError(undefined, undefined, 'ECONNABORTED')
    expect(getErrorMessage(err)).toBe('Koneksi timeout. Silakan coba lagi.')
  })

  it('maps network failure without response', () => {
    const err = makeAxiosError(undefined, undefined)
    expect(getErrorMessage(err)).toBe('Tidak dapat terhubung ke server. Periksa koneksi Anda.')
  })

  it('maps 401 to session expired', () => {
    const err = makeAxiosError(401, {})
    expect(getErrorMessage(err)).toBe('Sesi Anda telah berakhir. Silakan login kembali.')
  })

  it('maps 403 to access denied', () => {
    const err = makeAxiosError(403, {})
    expect(getErrorMessage(err)).toBe('Anda tidak memiliki akses untuk melakukan tindakan ini.')
  })

  it('maps 404 to not found', () => {
    const err = makeAxiosError(404, {})
    expect(getErrorMessage(err)).toBe('Data tidak ditemukan.')
  })

  it('maps 5xx to server error', () => {
    const err = makeAxiosError(500, {})
    expect(getErrorMessage(err)).toBe('Terjadi kesalahan pada server. Silakan coba lagi nanti.')
  })

  it('falls back to Error message for plain errors', () => {
    expect(getErrorMessage(new Error('Custom failure'))).toBe('Custom failure')
  })

  it('uses the fallback for unknown values', () => {
    expect(getErrorMessage('garbage')).toBe('Terjadi kesalahan. Silakan coba lagi.')
    expect(getErrorMessage(null, 'Kustom')).toBe('Kustom')
  })
})

describe('mapAuthError', () => {
  it('maps locked account', () => {
    expect(mapAuthError('Akun terkunci')).toBe(
      'Akun Anda terkunci. Silakan hubungi admin fakultas.',
    )
  })

  it('maps must-change-password', () => {
    expect(mapAuthError('Anda harus mengubah password')).toContain('mengganti password')
  })

  it('maps invalid credentials', () => {
    expect(mapAuthError('Email atau password salah')).toBe('Email atau password salah.')
  })

  it('passes through unknown messages', () => {
    expect(mapAuthError('Pesan lain')).toBe('Pesan lain')
  })
})
