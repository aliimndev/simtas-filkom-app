import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import LoginPage from '@/app/(auth)/login/page'

// next/navigation hooks require the Next.js router context; stub them for the
// smoke test so the client component can render in jsdom.
jest.mock('next/navigation', () => ({
  useRouter: () => ({ replace: jest.fn(), push: jest.fn() }),
  useSearchParams: () => ({ get: () => null }),
}))

// useLoginMutation talks to TanStack Query + the API client; return a fake so
// the page renders and submits without a network call.
jest.mock('@/lib/hooks/use-auth', () => ({
  useLoginMutation: () => ({
    mutateAsync: jest.fn().mockResolvedValue({}),
    isPending: false,
  }),
}))

describe('Login page (smoke)', () => {
  it('renders the login form fields', () => {
    render(<LoginPage />)
    expect(screen.getByLabelText('Email')).toBeInTheDocument()
    expect(screen.getByLabelText('Password')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: 'Masuk' })).toBeInTheDocument()
  })

  it('shows validation errors when submitting an empty form', async () => {
    const user = userEvent.setup()
    render(<LoginPage />)
    await user.click(screen.getByRole('button', { name: 'Masuk' }))
    expect(await screen.findByText('Email wajib diisi')).toBeInTheDocument()
    expect(await screen.findByText('Password wajib diisi')).toBeInTheDocument()
  })
})
